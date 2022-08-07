import React, { useContext } from "react";

export type Lang = "en" | "ja";

export interface LangState {
  lang: Lang;
}

export interface LangAction {
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

export const useLang = (): Lang => {
  const lang = useContext(LangContext);
  return lang?.state.lang || "en";
};
