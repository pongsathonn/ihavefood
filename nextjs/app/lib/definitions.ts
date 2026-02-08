
export type Role = "CUSTOMER" | "TODO"

export type SessionPayload = {
    userId: string;
    accessToken: string;
    role: Role;
};