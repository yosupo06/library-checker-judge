export type Timestamp = Date;

export enum SolvedStatus {
  UNKNOWN = "UNKNOWN",
  LATEST_AC = "LATEST_AC",
  AC = "AC",
}

export type User = {
  name: string;
  libraryUrl: string;
  isDeveloper: boolean;
};

export type CurrentUserInfoResponse = {
  user?: User;
};

export type ChangeCurrentUserInfoRequest = {
  user?: User;
};

export type UserInfoResponse = {
  isAdmin: boolean;
  user: User;
  solvedMap: Record<string, SolvedStatus>;
};

export type MonitoringResponse = {
  totalUsers: number;
  totalSubmissions: number;
  taskQueue: {
    pendingTasks: number;
    runningTasks: number;
    totalTasks: number;
  };
};

export type SubmissionOverview = {
  id: number;
  problemName: string;
  problemTitle: string;
  userName: string;
  lang: string;
  isLatest: boolean;
  status: string;
  hacked: boolean;
  time: number;
  memory: bigint;
  submissionTime?: Timestamp;
};

export type SubmissionCaseResult = {
  case: string;
  status: string;
  time: number;
  memory: bigint;
  stderr: Uint8Array;
  checkerOut: Uint8Array;
};

export type SubmissionListResponse = {
  submissions: SubmissionOverview[];
  count: number;
};

export type SubmissionInfoResponse = {
  overview: SubmissionOverview;
  caseResults: SubmissionCaseResult[];
  source: string;
  compileError: Uint8Array;
  canRejudge: boolean;
};

export type SubmitRequest = {
  problem: string;
  source: string;
  lang: string;
  tleKnockout?: boolean;
};

export type SubmitResponse = {
  id: number;
};

export type RejudgeRequest = {
  id: number;
};

export type RejudgeResponse = Record<string, never>;

export type HackRequest = {
  submission: number;
  testCase:
    | { oneofKind: "txt"; txt: Uint8Array }
    | { oneofKind: "cpp"; cpp: Uint8Array };
};

export type HackResponse = {
  id: number;
};

export type HackOverview = {
  id: number;
  submissionId: number;
  status: string;
  userName?: string;
  time?: number;
  memory?: bigint;
  hackTime: Timestamp;
};

export type HackListResponse = {
  hacks: HackOverview[];
  count: number;
};

export type HackInfoResponse = {
  overview: HackOverview;
  testCase:
    | { oneofKind: undefined }
    | { oneofKind: "txt"; txt: Uint8Array }
    | { oneofKind: "cpp"; cpp: Uint8Array };
  stderr?: Uint8Array;
  judgeOutput?: Uint8Array;
};

export type ProblemCategory = {
  title: string;
  problems: string[];
};
