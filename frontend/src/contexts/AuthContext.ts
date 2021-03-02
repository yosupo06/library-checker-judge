import React from "react";

export interface AuthState {
  user: string;
  token: string;
}

interface AuthLoginAction {
  type: "login";
  payload: AuthState;
}
interface AuthLogoutAction {
  type: "logout";
}

export type AuthAction = AuthLoginAction | AuthLogoutAction;

export const AuthReducer: React.Reducer<AuthState, AuthAction> = (
  state,
  action
) => {
  switch (action.type) {
    case "login":
      return action.payload;
    case "logout":
      return {
        user: "",
        token: "",
      };
    default:
      return state;
  }
};

export const AuthContext = React.createContext<{
  state: AuthState;
  dispatch: React.Dispatch<AuthAction>;
} | null>(null);
