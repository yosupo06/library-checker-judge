import { Problem, ProblemCategory } from "../api/library_checker_pb";

export const getCategories = (
  problems: Problem[],
  categories: ProblemCategory[]
): {
  name: string;
  problems: {
    name: string;
    title: string;
  }[];
}[] => {
  const classifiedSet = new Set(
    categories.map((e) => e.getProblemsList()).flat()
  );
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
    name: e.getTitle(),
    problems: e
      .getProblemsList()
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
