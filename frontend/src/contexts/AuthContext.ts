import React from 'react';

export interface AuthState {
  user: string
  token: string
}

export interface AuthAction {
  type: 'login'
  payload: AuthState
}

export const AuthReducer: React.Reducer<AuthState, AuthAction> = (state, action) => {
  console.log(state, action)
  switch (action.type) {
  case 'login':
    return action.payload
  default:
    return state
  }
}

export const AuthContext = React.createContext<{ state: AuthState; dispatch: React.Dispatch<AuthAction>; } | null>(null);
