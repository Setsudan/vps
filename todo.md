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
| 🔴 **P0**  | **Avatar upload → "Access Denied"**                               | Bucket is public‑read only; API needs `s3:PutObject`.<br/>  `mc anonymous set public local/$STORAGE_BUCKET` **or**<br/>  create a write‑enabled service user and use those creds in `go_launay`.          |                           |
| 🟠 **P1**  | **Categories ↔ Channels relation**                                | Current CRUD treats them separately. <br/>  `CategoryService.Create` should accept an optional slice of channels. <br/>  `ChannelService.Create` should require a `category_id` (nullable for top‑level). |
| 🟡 **P2**  | **Messaging / conversations untested**                            | Need manual / Postman flow: WS connect, send, reaction, Redis TTL → PG transfer.                                                                                                                              |
| 🟡 **P2**  | Migrate legacy "Groups" to new Guilds (optional)                  | Decide if legacy stays or is deprecated.                                                                                                                                                                      |
| 🟡 **P2**  | Swagger / tests                                                   | OpenAPI spec + unit/integration tests.                                                                                                                                                                        |
