import { PropsWithChildren, createContext, useContext, useEffect, useState } from "react";
import { User, UserContext } from "src/domain/UserContext";

// Create a context for the API response details
// eslint-disable-next-line import/no-default-export
export const UserDetailsContext = createContext<UserContext>({
  fullName: "",
  gmail: "",
  password: "",
  userDID: "",
  username: "",
  userType: "",
  // eslint-disable-next-line
  setUserDetails: () => {},
});

export function UserDetailsProvider(props: PropsWithChildren) {
  /* eslint-disable */

  const [userDetails, setUserDetailsState] = useState<User>(() => {
    // Initialize state from localStorage if available
    const storedDetails = localStorage.getItem("userDetails");
    return storedDetails
      ? JSON.parse(storedDetails)
      : {
          fullName: "",
          gmail: "",
          password: "",
          userDID: "",
          username: "",
          userType: "",
        };
  });

  /* eslint-enable */

  const setUserDetails = (details: User) => {
    setUserDetailsState(details);
    // Store the details in localStorage
    localStorage.setItem("userDetails", JSON.stringify(details));
  };

  useEffect(() => {
    // Add event listener to clear the storage on logout or other conditions
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === "userDetails" && e.newValue === null) {
        setUserDetailsState({
          fullName: "",
          gmail: "",
          password: "",
          userDID: "",
          username: "",
          userType: "",
        });
      }
    };

    window.addEventListener("storage", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, []);

  const contextValue = {
    fullName: userDetails.fullName,
    gmail: userDetails.gmail,
    password: userDetails.password,
    setUserDetails: (
      username: string,
      password: string,
      userDID: string,
      fullName: string,
      gmail: string,
      userType: string
    ): void => {
      setUserDetails({
        fullName,
        gmail,
        password,
        userDID,
        username,
        userType,
      });
    },
    userDID: userDetails.userDID,
    username: userDetails.username,
    userType: userDetails.userType,
  };

  return <UserDetailsContext.Provider value={contextValue} {...props} />;
}

export function useUserContext(): UserContext {
  return useContext(UserDetailsContext);
}
