export interface IUser {
  id: string;
  username: string;
  bio: string;
  email: string;
  role: string;
  avatar: string;
  status: Status;
  lastSeenAt?: Date;
  createdAt: Date;
  updatedAt: Date;
  guilds: string[];
}

export type Status = 'online' | 'offline' | 'away' | 'busy';
