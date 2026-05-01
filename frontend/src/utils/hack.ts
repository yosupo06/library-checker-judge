export const refactorTestCase = (testCase: string): string => {
  testCase = testCase.replaceAll("\r\n", "\n");
  if (testCase.length === 0 || testCase.slice(-1) != "\n") {
    testCase += "\n";
  }
  return testCase;
};
