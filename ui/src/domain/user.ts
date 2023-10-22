export type UserDetails = {
  PAN: string;
  address: string;
  adhar: string;
  createdAt: Date;
  dob: string;
  documentationSource: string;
  gmail: string;
  gstin: string;
  id: string;
  iscompleted: boolean;
  name: string;
  owner: string;
  phoneNumber: string;
  userType: string;
  username: string;
};

export type userProfile = {
  Address: string;
  Adhar: string;
  DOB: Date;
  DocumentationSource: string;
  Gmail: string;
  Gstin: string;
  ID: string;
  Name: string;
  Owner: string;
  PAN: string;
  PhoneNumber: string;
};
export type FormValue = {
  Aadhar: string;
  Age: string;
  PAN: string;
  address: string;
  dob: string;
  gst: string;
  mobile: string;
  owner: string;
  request: string;
};
export type FormData = {
  adhaarID: string;
  age: string;
  schemaID: string;
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

export type UserResponse = {
  msg: string;
  status: boolean;
};
