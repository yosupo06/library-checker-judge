import {
  useMutation,
  UseMutationOptions,
  useQuery,
  useQueryClient,
  UseQueryOptions,
  UseQueryResult,
} from "@tanstack/react-query";
import { LibraryCheckerServiceClient } from "../proto/library_checker.client";
import { GrpcWebFetchTransport } from "@protobuf-ts/grpcweb-transport";
import {
  ChangeCurrentUserInfoRequest,
  CurrentUserInfoResponse,
  HackInfoResponse,
  HackListResponse,
  HackRequest,
  HackResponse,
  MonitoringResponse,
  RejudgeRequest,
  SubmissionCaseResult,
  SubmissionInfoResponse,
  SubmissionListResponse,
  SubmissionOverview,
  SubmitRequest,
  SubmitResponse,
  UserInfoResponse,
} from "../proto/library_checker";
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
  fetchSubmissionList,
  fetchSubmissionInfo,
  postSubmit,
} from "./http_client";
import type { components as OpenApi } from "../openapi/types";
import { Timestamp as TimestampMessage } from "../proto/google/protobuf/timestamp";

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

const transport = new GrpcWebFetchTransport({
  baseUrl: import.meta.env.VITE_API_URL,
});
const client = new LibraryCheckerServiceClient(transport);
export default client;

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
    queryFn: async () => await client.monitoring({}, {}).response,
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
      const r = await fetchUserInfoREST(name);
      return {
        user: {
          name: r.user.name,
          libraryUrl: r.user.library_url,
          isDeveloper: r.user.is_developer,
        },
        solvedMap: (r.solved_map ?? {}) as Record<string, never>,
      } as unknown as UserInfoResponse;
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
        order === "-id" || order === "+time" ? order : undefined;
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
  const bearer = useBearer();
  return useMutation({
    mutationFn: async (req: RejudgeRequest) =>
      await client.rejudge(req, bearer.data ?? undefined).response,
  });
};

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
      ? TimestampMessage.fromDate(new Date(overview.submission_time))
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

const useBearer = () => {
  const idToken = useIdToken();
  return useQuery({
    queryKey: ["api", "bearer", idToken.data],
    queryFn: () => {
      return idToken.data
        ? {
            meta: {
              authorization: "bearer " + idToken.data,
            },
          }
        : null;
    },
    enabled: !idToken.isLoading,
  });
};

export const useHackMutation = (
  options?: Omit<
    UseMutationOptions<HackResponse, Error, HackRequest>,
    "mutationFn"
  >,
) => {
  const bearer = useBearer();
  return useMutation({
    mutationFn: async (req: HackRequest) =>
      await client.hack(req, bearer.data ?? undefined).response,
    ...options,
  });
};

export const useHackInfo = (
  id: number,
  options?: Omit<UseQueryOptions<HackInfoResponse>, "queryKey" | "queryFn">,
): UseQueryResult<HackInfoResponse> => {
  const bearer = useBearer();
  return useQuery({
    queryKey: ["hackInfo", String(id)],
    queryFn: async () =>
      await client.hackInfo(
        {
          id: id,
        },
        bearer.data ?? undefined,
      ).response,
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
  const bearer = useBearer();
  return useQuery({
    queryKey: ["hackList", user, status, order, skip, limit],
    queryFn: async () =>
      await client.hackList(
        {
          user: user,
          status: status,
          order: order,
          skip: skip,
          limit: limit,
        },
        bearer.data ?? undefined,
      ).response,
    structuralSharing: false,
  });
};
