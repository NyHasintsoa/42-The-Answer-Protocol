_This project has been created as part of the 42 curriculum by primary-user, secondary-user._

# The Answer Protocol

## 📖 Description

The Go server separates gameplay responsibilities into services in
[`server/src/service/`](./server/src/service/): `WorldManager`,
`InventoryManager`, `ChatManager`, `QuestManager`, and `NpcManager`.
`ActionRouter` delegates domain operations to these managers, while the
network/session and persistence services remain separate.

## 🛠️ Instructions

### Prerequisites

### Installation & Execution

The server stores its data in separate JSON files under `server/data/`:

- `players.json` stores each player's saved state, including room, health,
  inventory, and quests.
- `chat.json` stores chat messages.
- `interactions.json` stores non-chat interactions and server events.

An existing combined `interactions.json` database is split into these files
automatically the next time the server saves data.

## 📚 Resources

- **Graphics & Sprites:** Sourced from [CraftPix.net](https://craftpix.net/).
  - **Asset Packs Used:** [Name of Asset Pack](https://craftpix.net/link-to-exact-pack) _(e.g., 2D Game Tileset)_
  - **Usage:** Overworld tiles, UI icons, and character sprite sheets.
  - **License:** [CraftPix License](https://craftpix.net/file-licenses/) (Free/Premium Royalty-Free)

### References
