export type UserContext = {
  fullName: string;
  gmail: string;
  password: string;
  setUserDetails: (
    username: string,
    password: string,
    userDID: string,
    fullName: string,
    gmail: string,
    userType: string
  ) => void;
  userDID: string;
  userType: string;
  username: string;
};
export type User = {
  fullName: string;
  gmail: string;
  password: string;
  userDID: string;
  userType: string;
  username: string;
};
