import { Lang } from "../contexts/LangContext";

// UI element translations following the same pattern as statement parser
export const uiTranslations: { [key: string]: { [lang in Lang]: string } } = {
  // TLE Knockout feature
  tleKnockoutLabel: {
    en: "TLE Knockout",
    ja: "TLEノックアウト",
  },
  
  // Common UI elements
  submit: {
    en: "Submit",
    ja: "提出",
  },
  
  language: {
    en: "Language",
    ja: "言語",
  },
  
  languageLabel: {
    en: "Language:",
    ja: "言語:",
  },
  
  // Navigation (for future use)
  submissions: {
    en: "Submissions", 
    ja: "提出",
  },
  
  ranking: {
    en: "Ranking",
    ja: "ランキング",
  },
  
  help: {
    en: "Help",
    ja: "ヘルプ",
  },
};

// Hook to get translated text
export const useTranslation = (lang: Lang) => {
  return (key: string): string => {
    return uiTranslations[key]?.[lang] || key;
  };
};