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
  HackInfoResponse,
  HackListResponse,
  HackRequest,
  HackResponse,
  MonitoringResponse,
  RejudgeRequest,
  SubmissionInfoResponse,
  SubmissionListResponse,
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
} from "./http_client";
import type { components as OpenApi } from "../openapi/types";

const currentUserKey = ["api", "currentUser"];
export const useCurrentUser = () => {
  const bearer = useBearer();
  return useQuery({
    queryKey: ["api", "currentUser", bearer.data],
    queryFn: async () =>
      await client.currentUserInfo({}, bearer.data ?? undefined).response,
    enabled: !bearer.isLoading,
  });
};
export const useRegister = () => {
  const bearer = useBearer();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (name: string) =>
      await client.register(
        {
          name: name,
        },
        bearer.data ?? undefined,
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: currentUserKey,
      });
    },
  });
};
export const useChangeCurrentUserInfoMutation = () => {
  const bearer = useBearer();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (req: ChangeCurrentUserInfoRequest) =>
      await client.changeCurrentUserInfo(req, bearer.data ?? undefined)
        .response,
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
  const bearer = useBearer();
  return useQuery({
    queryKey: ["api", "userInfo", name, bearer.data?.meta.authorization ?? ""],
    queryFn: async () =>
      await client.userInfo({ name: name }, bearer.data ?? undefined).response,
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
    queryFn: async () =>
      await client.submissionList(
        {
          problem: problem,
          user: user,
          dedupUser: dedupUser,
          status: status,
          lang: lang,
          order: order,
          skip: skip,
          limit: limit,
          hacked: false,
        },
        {},
      ).response,
    structuralSharing: false,
  });

export const useSubmissionInfo = (
  id: number,
  options?: Omit<
    UseQueryOptions<SubmissionInfoResponse>,
    "queryKey" | "queryFn"
  >,
): UseQueryResult<SubmissionInfoResponse> => {
  const bearer = useBearer();
  return useQuery({
    queryKey: ["submissionInfo", String(id)],
    queryFn: async () =>
      await client.submissionInfo(
        {
          id: id,
        },
        bearer.data ?? undefined,
      ).response,
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
  const bearer = useBearer();
  return useMutation({
    mutationFn: async (req: SubmitRequest) =>
      await client.submit(req, bearer.data ?? undefined).response,
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
