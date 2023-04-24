import { Problem, ProblemCategory } from "../api/library_checker";

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
      dict[problem.name] = problem;
      return dict;
    },
    {}
  );

  const problemNames = problems.map((e) => e.name);
  const problemNameSet = new Set(problemNames);
  const classifiedSet = new Set(categories.map((e) => e.problems).flat());
  const newProblems = problemNames.filter((e) => !classifiedSet.has(e));

  const result = categories.map((category) => ({
    name: category.title,
    problems: category.problems
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
