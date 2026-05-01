import {
  useMutation,
  UseMutationOptions,
  useQuery,
  useQueryClient,
  UseQueryOptions,
  UseQueryResult,
} from "@tanstack/react-query";
import {
  ChangeCurrentUserInfoRequest,
  CurrentUserInfoResponse,
  HackInfoResponse,
  HackListResponse,
  HackRequest,
  HackResponse,
  HackOverview,
  MonitoringResponse,
  RejudgeRequest,
  RejudgeResponse,
  SubmissionCaseResult,
  SubmissionInfoResponse,
  SubmissionListResponse,
  SubmissionOverview,
  SubmitRequest,
  SubmitResponse,
  SolvedStatus,
  UserInfoResponse,
} from "./types";
import { useIdToken } from "../auth/auth";
import {
  fetchRanking,
  fetchProblemInfo,
  fetchProblemList,
  fetchLangList,
  fetchProblemCategories,
  fetchCurrentUserInfo,
  registerUser,
  patchCurrentUserInfo,
  fetchUserInfo as fetchUserInfoREST,
  fetchUserStatistics,
  fetchSubmissionList,
  fetchSubmissionInfo,
  postSubmit,
  fetchHackInfo,
  fetchHackList,
  postHack,
  fetchMonitoring,
  postRejudge,
} from "./http_client";
import type { components as OpenApi } from "../openapi/types";

const currentUserKey = ["api", "currentUser"];
export const useCurrentUser = () => {
  const idToken = useIdToken();
  return useQuery<CurrentUserInfoResponse>({
    queryKey: ["api", "currentUser", idToken.data ?? ""],
    queryFn: async () => {
      const r = await fetchCurrentUserInfo(idToken.data ?? undefined);
      const user = r.user
        ? {
            name: r.user.name,
            libraryUrl: r.user.library_url,
            isDeveloper: r.user.is_developer,
          }
        : undefined;
      return { user } as CurrentUserInfoResponse;
    },
    enabled: !idToken.isLoading,
  });
};
export const useRegister = () => {
  const idToken = useIdToken();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (name: string) =>
      await registerUser(name, idToken.data ?? undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: currentUserKey,
      });
    },
  });
};
export const useChangeCurrentUserInfoMutation = () => {
  const idToken = useIdToken();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (req: ChangeCurrentUserInfoRequest) =>
      await patchCurrentUserInfo(
        {
          name: req.user!.name,
          library_url: req.user!.libraryUrl,
          is_developer: req.user!.isDeveloper,
        },
        idToken.data ?? undefined,
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: currentUserKey,
      });
    },
  });
};

export const useLangList = (): UseQueryResult<
  OpenApi["schemas"]["LangListResponse"]
> =>
  useQuery({
    queryKey: ["langList"],
    // Use REST endpoint
    queryFn: async () => await fetchLangList(),
  });

export const useRanking = (
  skip: number = 0,
  limit: number = 100,
): UseQueryResult<OpenApi["schemas"]["RankingResponse"]> =>
  useQuery({
    queryKey: ["ranking", skip, limit],
    // Use REST for migrated endpoint
    queryFn: async () => await fetchRanking(skip, limit),
  });

export const useMonitoring = (): UseQueryResult<MonitoringResponse> =>
  useQuery({
    queryKey: ["monitoring"],
    queryFn: async () => {
      const res = await fetchMonitoring();
      return toMonitoringProto(res);
    },
    refetchInterval: 30000, // Refetch every 30 seconds for real-time monitoring
  });

export const useProblemInfo = (
  name: string,
): UseQueryResult<OpenApi["schemas"]["ProblemInfoResponse"]> =>
  useQuery({
    queryKey: ["problemInfo", name],
    // Use REST endpoint
    queryFn: async () => await fetchProblemInfo(name),
    structuralSharing: false,
  });

export const useProblemList = (): UseQueryResult<
  OpenApi["schemas"]["ProblemListResponse"]
> =>
  useQuery({
    queryKey: ["problemList"],
    // Use REST endpoint
    queryFn: async () => await fetchProblemList(),
  });

export const useProblemCategories = (): UseQueryResult<
  OpenApi["schemas"]["ProblemCategoriesResponse"]
> =>
  useQuery({
    queryKey: ["problemCategories"],
    // Use REST endpoint
    queryFn: async () => await fetchProblemCategories(),
  });

export const useUserInfo = (
  name: string,
  options?: Omit<UseQueryOptions<UserInfoResponse>, "queryKey" | "queryFn">,
): UseQueryResult<UserInfoResponse> => {
  return useQuery({
    queryKey: ["api", "userInfo", name],
    queryFn: async () => {
      const [userInfo, stats] = await Promise.all([
        fetchUserInfoREST(name),
        fetchUserStatistics(name),
      ]);
      const solvedMap = toSolvedStatusMap(stats.solved_map ?? {});
      return {
        isAdmin: false,
        user: {
          name: userInfo.user.name,
          libraryUrl: userInfo.user.library_url,
          isDeveloper: userInfo.user.is_developer,
        },
        solvedMap,
      } satisfies UserInfoResponse;
    },
    ...options,
  });
};

export const useSubmissionList = (
  problem: string,
  user: string,
  dedupUser: boolean,
  status: string,
  lang: string,
  order: string,
  skip: number,
  limit: number,
): UseQueryResult<SubmissionListResponse> =>
  useQuery({
    queryKey: [
      "submissionList",
      problem,
      user,
      dedupUser,
      status,
      lang,
      order,
      skip,
      limit,
    ],
    // Use REST endpoint
    queryFn: async () => {
      const orderParam =
        order === "-id" || order === "+time"
          ? (order as OpenApi["schemas"]["SubmissionOrder"])
          : undefined;
      const res = await fetchSubmissionList({
        problem,
        user,
        dedupUser,
        status,
        lang,
        order: orderParam,
        skip,
        limit,
      });
      return {
        submissions: res.submissions.map(toSubmissionOverviewProto),
        count: res.count,
      } satisfies SubmissionListResponse;
    },
    structuralSharing: false,
  });

export const useSubmissionInfo = (
  id: number,
  options?: Omit<
    UseQueryOptions<SubmissionInfoResponse>,
    "queryKey" | "queryFn"
  >,
): UseQueryResult<SubmissionInfoResponse> => {
  return useQuery({
    queryKey: ["submissionInfo", String(id)],
    queryFn: async () => {
      const res = await fetchSubmissionInfo(id);
      return toSubmissionInfoProto(res);
    },
    structuralSharing: false,
    ...options,
  });
};

export const useSubmitMutation = (
  options?: Omit<
    UseMutationOptions<SubmitResponse, Error, SubmitRequest>,
    "mutationFn"
  >,
) => {
  const idToken = useIdToken();
  return useMutation({
    mutationFn: async (req: SubmitRequest) => {
      const res = await postSubmit(
        {
          problem: req.problem,
          source: req.source,
          lang: req.lang,
          tleKnockout: req.tleKnockout,
        },
        idToken.data ?? undefined,
      );
      return { id: res.id } satisfies SubmitResponse;
    },
    ...options,
  });
};

export const useRejudgeMutation = () => {
  const idToken = useIdToken();
  return useMutation<RejudgeResponse, Error, RejudgeRequest>({
    mutationFn: async (req: RejudgeRequest) => {
      await postRejudge(req.id, idToken.data ?? undefined);
      return {};
    },
  });
};

const toMonitoringProto = (
  res: OpenApi["schemas"]["MonitoringResponse"],
): MonitoringResponse => ({
  totalUsers: res.total_users,
  totalSubmissions: res.total_submissions,
  taskQueue: {
    pendingTasks: res.task_queue.pending_tasks,
    runningTasks: res.task_queue.running_tasks,
    totalTasks: res.task_queue.total_tasks,
  },
});

const toSubmissionInfoProto = (
  res: OpenApi["schemas"]["SubmissionInfoResponse"],
): SubmissionInfoResponse => {
  const overview = toSubmissionOverviewProto(res.overview);
  const caseResults = res.case_results?.map(toSubmissionCaseResultProto) ?? [];
  return {
    overview,
    caseResults,
    source: res.source,
    compileError: decodeBase64(res.compile_error),
    canRejudge: res.can_rejudge,
  } satisfies SubmissionInfoResponse;
};

const toSubmissionOverviewProto = (
  overview: OpenApi["schemas"]["SubmissionOverview"],
): SubmissionOverview => {
  return {
    id: overview.id,
    problemName: overview.problem_name,
    problemTitle: overview.problem_title,
    userName: overview.user_name ?? "",
    lang: overview.lang,
    isLatest: overview.is_latest,
    status: overview.status,
    hacked: false,
    time: overview.time,
    memory: BigInt(overview.memory),
    submissionTime: overview.submission_time
      ? new Date(overview.submission_time)
      : undefined,
  } satisfies SubmissionOverview;
};

const toSubmissionCaseResultProto = (
  res: OpenApi["schemas"]["SubmissionCaseResult"],
): SubmissionCaseResult => {
  return {
    case: res.case,
    status: res.status,
    time: res.time,
    memory: BigInt(res.memory),
    stderr: decodeBase64(res.stderr),
    checkerOut: decodeBase64(res.checker_out),
  } satisfies SubmissionCaseResult;
};

const decodeBase64 = (value?: string | null): Uint8Array => {
  if (!value) {
    return new Uint8Array();
  }
  const globalAtob = (
    globalThis as {
      atob?: (data: string) => string;
    }
  ).atob;
  if (typeof globalAtob === "function") {
    const binary = globalAtob(value);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }
  const globalBuffer = (
    globalThis as {
      Buffer?: { from: (data: string, encoding: string) => Uint8Array };
    }
  ).Buffer;
  if (globalBuffer) {
    return globalBuffer.from(value, "base64");
  }
  throw new Error("Base64 decoder is not available in this environment");
};

const toSolvedStatusMap = (
  solvedMap: Record<string, OpenApi["schemas"]["SolvedStatus"]>,
): Record<string, SolvedStatus> => {
  const mapped: Record<string, SolvedStatus> = {};
  Object.entries(solvedMap).forEach(([problem, status]) => {
    switch (status) {
      case "LATEST_AC":
        mapped[problem] = SolvedStatus.LATEST_AC;
        break;
      case "AC":
        mapped[problem] = SolvedStatus.AC;
        break;
      default:
        mapped[problem] = SolvedStatus.UNKNOWN;
    }
  });
  return mapped;
};

export const useHackMutation = (
  options?: Omit<
    UseMutationOptions<HackResponse, Error, HackRequest>,
    "mutationFn"
  >,
) => {
  const idToken = useIdToken();
  return useMutation({
    mutationFn: async (req: HackRequest) => {
      const payload = {
        submission: req.submission,
        testCaseTxt:
          req.testCase.oneofKind === "txt" ? req.testCase.txt : undefined,
        testCaseCpp:
          req.testCase.oneofKind === "cpp" ? req.testCase.cpp : undefined,
      };
      const res = await postHack(payload, idToken.data ?? undefined);
      return toHackResponseProto(res);
    },
    ...options,
  });
};

export const useHackInfo = (
  id: number,
  options?: Omit<UseQueryOptions<HackInfoResponse>, "queryKey" | "queryFn">,
): UseQueryResult<HackInfoResponse> => {
  return useQuery({
    queryKey: ["hackInfo", String(id)],
    queryFn: async () => {
      const res = await fetchHackInfo(id);
      return toHackInfoProto(res);
    },
    ...options,
  });
};

export const useHackList = (
  user: string,
  status: string,
  order: string,
  skip: number,
  limit: number,
): UseQueryResult<HackListResponse> => {
  return useQuery({
    queryKey: ["hackList", user, status, order, skip, limit],
    queryFn: async () => {
      const res = await fetchHackList({ user, status, order, skip, limit });
      return {
        hacks: res.hacks.map(toHackOverviewProto),
        count: res.count,
      } satisfies HackListResponse;
    },
    structuralSharing: false,
  });
};

const toHackResponseProto = (
  res: OpenApi["schemas"]["HackResponse"],
): HackResponse => ({ id: res.id });

const toHackOverviewProto = (
  overview: OpenApi["schemas"]["HackOverview"],
): HackOverview => ({
  id: overview.id,
  submissionId: overview.submission_id,
  status: overview.status,
  userName: overview.user_name,
  time: overview.time,
  memory: overview.memory !== undefined ? BigInt(overview.memory) : undefined,
  hackTime: new Date(overview.hack_time),
});

const toHackInfoProto = (
  res: OpenApi["schemas"]["HackInfoResponse"],
): HackInfoResponse => {
  let testCase: HackInfoResponse["testCase"] = { oneofKind: undefined };
  if (res.test_case_txt) {
    testCase = { oneofKind: "txt", txt: decodeBase64(res.test_case_txt) };
  } else if (res.test_case_cpp) {
    testCase = { oneofKind: "cpp", cpp: decodeBase64(res.test_case_cpp) };
  }
  return {
    overview: toHackOverviewProto(res.overview),
    testCase,
    stderr: res.stderr ? decodeBase64(res.stderr) : undefined,
    judgeOutput: res.judge_output ? decodeBase64(res.judge_output) : undefined,
  } satisfies HackInfoResponse;
};
