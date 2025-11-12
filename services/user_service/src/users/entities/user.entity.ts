import { UserPreference } from "./user-preference.entity";

export class User {
  id: string;
  email: string;
  name?: string;
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export class UserWithPreferences extends User {
  preferences?: UserPreference;
}