import { PropsWithChildren, createContext, useContext, useState } from "react";
import { UserContext } from "src/domain/UserContext";

// Create a context for the API response details
export const UserDetailsContext = createContext<UserContext>({
  password: "",
  UserDID: "",
  username: "",
  // setUserDetails: () => {},
});

export function UserDetailsProvider(props: PropsWithChildren) {
  const [userDetails, setUserDetails] = useState({
    password: "",
    UserDID: "",
    username: "",
  });

  const contextValue = {
    password: userDetails.password,
    setUserDetails: (username: string, password: string, UserDID: string): void => {
      setUserDetails({ password, UserDID, username });
    },
    UserDID: userDetails.UserDID,
    username: userDetails.username,
  };

  return <UserDetailsContext.Provider value={contextValue} {...props} />;
}

export function useUserContext() {
  return useContext(UserDetailsContext);
}
