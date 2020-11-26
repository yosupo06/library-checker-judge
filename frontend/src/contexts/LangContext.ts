import React from "react";

type Lang = "en" | "ja";

interface LangState {
  lang: Lang;
}

interface LangAction {
  type: "change";
  payload: Lang;
}

export const LangReducer: React.Reducer<LangState, LangAction> = (
  state,
  action
) => {
  console.log(state, action);
  switch (action.type) {
  case "change":
    return { lang: action.payload };
  default:
    return state;
  }
};

export const LangContext = React.createContext<{
  state: LangState;
  dispatch: React.Dispatch<LangAction>;
} | null>(null);
