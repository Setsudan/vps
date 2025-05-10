# Massive Rework — progress snapshot (06 May 2025)

## ✅ Done

| Area               | Details                                                                                                                                                                                                                                                          |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Models**         | Core `User`, `Message`; legacy `Group` + `GroupMembership`; friendship `FriendRequest`; guilds (`Guild`, `GuildRole`, `GuildMember`, `Category`, `Channel`, `PermissionOverwrite`); résumé aggregate + sub‑models                                                |
| **Repositories**   | User, Messaging, Group, GroupMembership, FriendRequest, Guild, GuildMember, Category, Channel, PermissionOverwrite, Resume                                                                                                                                       |
| **Services**       | Auth · Users · Legacy Groups · Friendships · Messaging (Redis) · Presence · Résumés · Guilds · **Permissions** · **Categories** · **Channels** · **Guild‑Roles**                                                                                                 |
| **Controllers**    | Helpers · Auth · User · Messaging · Presence · Group (legacy) · Friendship · Resume · Guild · **Permissions** · **Categories** · **Channels** · **GuildRoles**                                                                                                   |
| **Boot & Routing** | `init_server.go` wires every new repo/service/controller; Redis TTL goroutine + expired‑key listener; Traefik stack rebuilt: <br>• API served under `/api/*` <br>• MinIO exposed at `/storage/*` <br>• Bucket initialised once via `minio-init` helper container |
| **Storage**        | `StorageService` simplified (no bucket ops); public read + write policy applied once by `minio-init`                                                                                                                                                             |

---

## 🪣 Open TODOS & Known Issues

|  Priority  | Item / Bug                                                        | Notes / Next step                                                                                                                                                                                             |
| ---------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 🟠 **P1**  | **Categories ↔ Channels relation**                                | Current CRUD treats them separately. <br/>  `CategoryService.Create` should accept an optional slice of channels. <br/>  `ChannelService.Create` should require a `category_id` (nullable for top‑level). |
| 🟡 **P2**  | **Messaging / conversations untested**                            | Need manual / Postman flow: WS connect, send, reaction, Redis TTL → PG transfer.                                                                                                                              |
| 🟡 **P2**  | Swagger / tests                                                   | OpenAPI spec + unit/integration tests.                                                                                                                                                                        |
