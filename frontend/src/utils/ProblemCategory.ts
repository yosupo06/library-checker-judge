import { Problem } from "../api/library_checker_pb";
import categories from "./categories.json";

export const getCategories = (
  problems: Problem[]
): {
  name: string;
  problems: {
    name: string;
    title: string;
  }[];
}[] => {
  const classifiedSet = new Set(categories.map((e) => e.problems).flat());
  const problemNames = problems.map((e) => e.getName());
  const nameToTitle = problems.reduce<{ [name: string]: string }>(
    (dict, problem) => {
      dict[problem.getName()] = problem.getTitle();
      return dict;
    },
    {}
  );

  const problemNameSet = new Set(problemNames);
  const classified = categories.map((e) => ({
    name: e.name,
    problems: e.problems
      .filter((e) => problemNameSet.has(e))
      .map((e) => ({
        name: e,
        title: nameToTitle[e],
      })),
  }));
  const unclassified = problemNames.filter((e) => !classifiedSet.has(e));
  if (unclassified.length) {
    classified.unshift({
      name: "New",
      problems: unclassified.map((e) => ({
        name: e,
        title: nameToTitle[e],
      })),
    });
  }
  return classified;
};
