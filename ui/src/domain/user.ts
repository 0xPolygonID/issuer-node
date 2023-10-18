export type UserDetails = {
  PAN: string;
  address: string;
  adhar: string;
  createdAt: Date;
  documentationSource: string;
  gmail: string;
  gstin: string;
  id: string;
  name: string;
  owner: string;
  userType: string;
  username: string;
};

export type Login = {
  fullName: string;
  gmail: string;
  iscompleted: boolean;
  password: string;
  userDID: string;
  userType: string;
  username: string;
};
