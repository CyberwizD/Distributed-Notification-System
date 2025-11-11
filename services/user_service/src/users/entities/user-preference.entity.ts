export class UserPreference {
  id: string;
  userId: string;
  emailEnabled: boolean;
  pushEnabled: boolean;
  smsEnabled: boolean;
  emailFrequency: string;
  language: string;
  timezone: string;
  marketingEmails: boolean;
  securityEmails: boolean;
  createdAt: Date;
  updatedAt: Date;
}