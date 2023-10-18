import { PropsWithChildren, createContext, useContext, useState } from "react";
import { User, UserContext } from "src/domain/UserContext";

// Create a context for the API response details
export const UserDetailsContext = createContext<UserContext>({
  fullName: "",
  gmail: "",
  password: "",
  UserDID: "",
  username: "",
  userType: "",
  // setUserDetails: () => {},
});

export function UserDetailsProvider(props: PropsWithChildren) {
  const [userDetails, setUserDetails] = useState<User>({
    fullName: "",
    gmail: "",
    password: "",
    UserDID: "",
    username: "",
    userType: "",
  });

  const contextValue = {
    fullName: userDetails.fullName,
    gmail: userDetails.gmail,
    password: userDetails.password,
    setUserDetails: (
      password: string,
      UserDID: string,
      username: string,
      fullName: string,
      gmail: string,
      userType: string
    ): void => {
      setUserDetails({ fullName, gmail, password, UserDID, username, userType });
    },
    UserDID: userDetails.UserDID,
    username: userDetails.username,
    userType: userDetails.userType,
  };

  return <UserDetailsContext.Provider value={contextValue} {...props} />;
}

export function useUserContext() {
  return useContext(UserDetailsContext);
}
