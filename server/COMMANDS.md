# Server Commands Reference

This document lists every command accepted by the TCP game server, with usage
examples and the responses the server returns.

## Protocol overview

- The client connects over TCP and sends one command per line, terminated by
  `\n`.
- Every command is answered with a single response line.
- Responses are prefixed with a status token:
  - `OK ...` — command succeeded.
  - `ERR <code> ...` — command failed (see [Error codes](#error-codes)).
- On connection, before any command is sent, the server emits:
  ```
  OK hello proto=1
  ```
- Commands are case-insensitive (they are upper-cased by the router).
- Arguments are separated by spaces. Several commands accept a single argument
  that may itself contain spaces (for example `CHAT` and `GROUP`).

All examples below show what the client sends (after `->`) and what the server
replies (after `<-`).

---

## 1. Connection commands

### CONNECT

Authenticate and join the server. This **must** be the first command; all other
commands require an authenticated player.

```
CONNECT <username>
```

- `<username>`: 1–32 characters, no whitespace or control characters.
- If a saved player exists for that username, their progress (room, HP,
  inventory, quests) is restored. Otherwise a new player is created in the
  `start` room with 100 HP.

**Example — new player**
```
-> CONNECT alice
<- OK connected
```

**Example — errors**
```
-> CONNECT
<- ERR 400 INVALID_USERNAME

-> CONNECT a_very_long_username_exceeding_thirty_two_characters_limit
<- ERR 400 INVALID_USERNAME

-> CONNECT bob
<- ERR 201 NAME_IN_USE        # bob is already logged in
```

### QUIT

Disconnect from the server. The connection is closed after the response.

```
QUIT
```

**Example**
```
-> QUIT
<- OK bye
```

---

## 2. World commands

### MAP

Return the full world map configuration as JSON (rooms, NPCs, items, quests).

```
MAP
```

**Example**
```
-> MAP
<- OK {"world":{"rooms":{"start":{"name":"Village Square",...},...},"npcs":{...},"items":{...},"quests":{...}}}
```

### LOOK

Describe the room the player is currently in, as JSON (exits, items, NPCs).

```
LOOK
```

**Example**
```
-> LOOK
<- OK {"id":"start","name":"Village Square","description":"A bustling square ...","exits":{"north":"alchemist_shop","east":"shop",...},"items":["potion_health_PM"],"npcs":["guard","mayor"]}
```

### MOVE

Move to an adjacent room through an exit.

```
MOVE <direction>
```

- `<direction>`: one of `north`, `south`, `east`, `west` (no spaces allowed).

**Example**
```
-> MOVE north
<- OK room=alchemist_shop
```

**Example — errors**
```
-> MOVE
<- ERR 400 INVALID_ARGUMENTS

-> MOVE up
<- ERR 301 NO_EXIT
```

### WHO

Return the number of players currently connected.

```
WHO
```

**Example**
```
-> WHO
<- OK players=3
```

---

## 3. Chat commands

### CHAT

Send a chat message to a scope.

```
CHAT <scope> <message>
```

- `<scope>`: `GLOBAL`, `ROOM`, or `GROUP`.
- `<message>`: the text to send.
- `GROUP` scope requires the player to be in a group.

The message is broadcast as an event line (`EVT ...`) to the relevant
recipients; the sender only receives the `OK` acknowledgement.

**Example — global**
```
-> CHAT GLOBAL Hello everyone!
<- OK
# All connected clients receive: EVT GLOBAL CHAT alice Hello everyone!
```

**Example — room**
```
-> CHAT ROOM Is anyone here?
<- OK
# Clients in the same room receive: EVT ROOM CHAT alice Is anyone here?
```

**Example — group**
```
-> CHAT GROUP Let's group up!
<- OK
# Group members receive: EVT GROUP CHAT alice Let's group up!
```

**Example — errors**
```
-> CHAT
<- ERR 400 INVALID_CHAT

-> CHAT GLOBAL
<- ERR 400 INVALID_CHAT

-> CHAT GROUP Meet me at the start
<- ERR 401 NOT_IN_GROUP

-> CHAT FOO bar
<- ERR 400 INVALID_SCOPE
```

---

## 4. Group commands

### GROUP

Manage player groups. Sub-commands: `CREATE`, `INVITE`, `JOIN`, `LEAVE`.

A group is identified by `grp_<leader>`.

```
GROUP <action> [target]
```

#### GROUP CREATE

Create a new group led by the player.

```
GROUP CREATE
```

**Example**
```
-> GROUP CREATE
<- OK group=grp_alice
```

#### GROUP INVITE

Invite another online player to the player's group. Only the group leader can
invite; the target must be online and not already in a group.

```
GROUP INVITE <player>
```

**Example**
```
-> GROUP INVITE bob
<- OK
# bob receives: EVT GROUP INVITE alice
```

#### GROUP JOIN

Join a group by leader name. Requires a pending invite from that leader.

```
GROUP JOIN <leader>
```

**Example**
```
-> GROUP JOIN alice
<- OK group=grp_alice
# Group members receive: EVT GROUP JOIN bob
```

#### GROUP LEAVE

Leave the current group.

```
GROUP LEAVE
```

**Example**
```
-> GROUP LEAVE
<- OK
# Former group members receive: EVT GROUP LEAVE bob
```

**Example — errors**
```
-> GROUP
<- ERR 400 INVALID_GROUP_CMD

-> GROUP CREATE extra
<- ERR 400 INVALID_ARGUMENTS

-> GROUP INVITE
<- ERR 400 INVALID_ARGUMENTS

-> GROUP INVITE carol
<- ERR 401 NOT_IN_GROUP          # sender is not in a group

-> GROUP JOIN dan
<- ERR 404 GROUP_NOT_FOUND        # no group or no invite

-> GROUP LEAVE extra
<- ERR 400 INVALID_ARGUMENTS
```

---

## 5. Inventory commands

### TAKE

Pick up an item from the current room.

```
TAKE <item>
```

**Example**
```
-> TAKE potion_health_PM
<- OK taken=potion_health_PM
```

### DROP

Drop an item from the player's inventory into the current room.

```
DROP <item>
```

**Example**
```
-> DROP potion_health_PM
<- OK dropped=potion_health_PM
```

### USE

Use an item from the inventory. Heal items restore HP; attack items add an
attack bonus.

```
USE <item>
```

**Example — heal**
```
-> USE potion_health_PM
<- OK {"item_id":"potion_health_PM","name":"Petite Health Potion","category":"heal","amount":25,"hp":75,"max_hp":100,"attack_bonus":0}
```

**Example — attack**
```
-> USE potion_attack_PM
<- OK {"item_id":"potion_attack_PM","name":"Petite Attack Potion","category":"attack","amount":25,"hp":100,"max_hp":100,"attack_bonus":25}
```

### INVENTORY

List the items in the player's inventory as JSON.

```
INVENTORY
```

**Example**
```
-> INVENTORY
<- OK ["dagger","potion_health_PM"]
```

---

## 6. NPC / Quest commands

### TALK

Talk to an NPC in the current room and receive its dialogue line.

```
TALK <npc>
```

**Example**
```
-> TALK guard
<- OK I used to be an adventurer like you, but then I realized guard duty has a dental plan.
```

### QUEST

Interact with a quest-giving NPC in the current room: accept an available
quest or turn in a completed one.

```
QUEST <npc>
```

**Example — quest available**
```
-> QUEST guard
<- OK {"quest_id":"fetch_herbs","description":"Bring 1 Healing Herb from the Hidden Grove to the Village Guard.","reward":"potion_health_PM","status":"available"}
```

**Example — quest completed**
```
-> QUEST guard
<- OK {"quest_id":"fetch_herbs","description":"...","reward":"potion_health_PM","status":"completed"}
```

### QUESTS

List all of the player's quests and their status as JSON.

```
QUESTS
```

**Example**
```
-> QUESTS
<- OK [{"quest_id":"fetch_herbs","status":"active","progress":"0/1"},{"quest_id":"fetch_amulet","status":"completed","progress":""}]
```

---

## 7. Combat commands

### ATTACK

Attack a hostile (attackable) NPC in the current room.

```
ATTACK <npc>
```

**Example**
```
-> ATTACK goblin
<- OK {"attacker_hp":90,"target_hp":15,"damage":15,"status":"combat"}
```

Possible `status` values: `combat`, `victory`, `respawned`.

### DEFEND

Take a defensive turn in combat, halving the incoming damage. Only works while
in combat (`INATTACK`).

```
DEFEND
```

**Example**
```
-> DEFEND
<- OK {"hp":85,"damage":2,"status":"defending"}
```

### FLEE

Attempt to flee combat in a given direction, taking full damage from the NPC.

```
FLEE <direction>
```

**Example**
```
-> FLEE south
<- OK {"hp":80,"damage":10,"status":"fled"}
```

### STATUS

Return the player's current combat status (HP, max HP, attack bonus).

```
STATUS
```

**Example**
```
-> STATUS
<- OK {"hp":100,"max_hp":100,"attack_bonus":0,"status":"healthy"}
```

During combat:
```
<- OK {"hp":75,"max_hp":100,"attack_bonus":0,"status":"combat"}
```

### NPC Defeat and Respawn

When a hostile NPC (one with `can_be_attacked: true`) is defeated in combat (its HP reaches 0), it is temporarily removed from the map. The NPC disappears from its room and cannot be targeted by `ATTACK`, `TALK`, or `QUEST` until it respawns.

Defeated NPCs respawn automatically after **5 minutes** with full HP restored to their original room.

**Example — defeating a goblin**
```
-> ATTACK goblin
<- OK {"attacker_hp":90,"target_hp":0,"damage":30,"status":"victory"}
# The goblin is removed from the room.
-> ATTACK goblin
<- ERR 404 NPC_NOT_FOUND
# After 5 minutes, the goblin respawns in forest_edge with full HP.
-> LOOK
# The goblin appears again in the room's NPC list.
```

---

## Error codes

| Code | Meaning |
|------|---------|
| `ERR 400 ...` | Bad request / invalid arguments / unknown command |
| `ERR 401 ...` | Not authenticated / not in group |
| `ERR 201 NAME_IN_USE` | Username already logged in |
| `ERR 301 NO_EXIT` | No exit in that direction |
| `ERR 404 ...` | Item / NPC / group not found |
| `ERR 402 ALREADY_IN_GROUP` | Player already belongs to a group |
| `ERR 405 ...` | Action not allowed (not hostile NPC, not in combat, item not usable) |
| `ERR 406 NO_QUEST_AVAILABLE` | NPC has no quest for the player |
| `ERR 409 HEALTH_FULL` | Player HP is already full |
| `ERR 413 MESSAGE_TOO_LONG` | Input line exceeds 1024 bytes |
| `ERR 422 INVALID_ITEM_EFFECT` | Item has no usable effect |
| `ERR 500 INTERNAL_ERROR` | Server-side failure |

---

## Full session example

```
# Connect
-> CONNECT alice
<- OK connected

# Look around
-> LOOK
<- OK {"id":"start","name":"Village Square",...}

# Move to the alchemist shop
-> MOVE east
<- OK room=shop

# Take an item
-> TAKE sword
<- OK taken=sword

# Check inventory
-> INVENTORY
<- OK ["sword"]

# Use a potion (must be in inventory first)
-> TAKE potion_health_PM
<- OK taken=potion_health_PM
-> USE potion_health_PM
<- OK {"item_id":"potion_health_PM","name":"Petite Health Potion","category":"heal","amount":25,"hp":100,"max_hp":100,"attack_bonus":0}

# Talk to the merchant
-> TALK merchant
<- OK Welcome! Buy something or please stop breathing on my shiny weapons.

# Interact with a quest NPC
-> QUEST guard
<- OK {"quest_id":"fetch_herbs","description":"Bring 1 Healing Herb ...","reward":"potion_health_PM","status":"available"}

# List quests
-> QUESTS
<- OK [{"quest_id":"fetch_herbs","status":"active","progress":"0/1"}]

# Check status
-> STATUS
<- OK {"hp":100,"max_hp":100,"attack_bonus":0,"status":"healthy"}

# Chat globally
-> CHAT GLOBAL heading to the hidden grove
<- OK

# Disconnect
-> QUIT
<- OK bye
```
