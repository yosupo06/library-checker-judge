import { Problem, ProblemCategory } from "../api/library_checker_pb";

export type CategorisedProblems = {
  name: string;
  problems: Problem[];
}[];

export const categoriseProblems = (
  problems: Problem[],
  categories: ProblemCategory[]
): CategorisedProblems => {
  const nameToProblem = problems.reduce<{ [name: string]: Problem }>(
    (dict, problem) => {
      dict[problem.getName()] = problem;
      return dict;
    },
    {}
  );

  const problemNames = problems.map((e) => e.getName());
  const problemNameSet = new Set(problemNames);
  const classifiedSet = new Set(
    categories.map((e) => e.getProblemsList()).flat()
  );
  const newProblems = problemNames.filter((e) => !classifiedSet.has(e));

  const result = categories.map((category) => ({
    name: category.getTitle(),
    problems: category
      .getProblemsList()
      .filter((e) => problemNameSet.has(e))
      .map((e) => nameToProblem[e]),
  }));

  if (newProblems.length) {
    result.unshift({
      name: "New",
      problems: newProblems.map((e) => nameToProblem[e]),
    });
  }
  return result;
};
