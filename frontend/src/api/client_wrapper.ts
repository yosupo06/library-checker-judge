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
  LangListResponse,
  MonitoringResponse,
  ProblemCategoriesResponse,
  ProblemInfoResponse,
  ProblemListResponse,
  RankingResponse,
  RejudgeRequest,
  SubmissionInfoResponse,
  SubmissionListResponse,
  SubmitRequest,
  SubmitResponse,
  UserInfoResponse,
} from "../proto/library_checker";
import { useIdToken } from "../auth/auth";

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

export const useLangList = (): UseQueryResult<LangListResponse> =>
  useQuery({
    queryKey: ["langList"],
    queryFn: async () => await client.langList({}, {}).response,
  });

export const useRanking = (
  skip: number = 0,
  limit: number = 100,
): UseQueryResult<RankingResponse> =>
  useQuery({
    queryKey: ["ranking", skip, limit],
    queryFn: async () => await client.ranking({ skip, limit }, {}).response,
  });

export const useMonitoring = (): UseQueryResult<MonitoringResponse> =>
  useQuery({
    queryKey: ["monitoring"],
    queryFn: async () => await client.monitoring({}, {}).response,
    refetchInterval: 30000, // Refetch every 30 seconds for real-time monitoring
  });

export const useProblemInfo = (
  name: string,
): UseQueryResult<ProblemInfoResponse> =>
  useQuery({
    queryKey: ["problemInfo", name],
    queryFn: async () => await client.problemInfo({ name: name }, {}).response,
    structuralSharing: false,
  });

export const useProblemList = (): UseQueryResult<ProblemListResponse> =>
  useQuery({
    queryKey: ["problemList"],
    queryFn: async () => await client.problemList({}, {}).response,
  });

export const useProblemCategories =
  (): UseQueryResult<ProblemCategoriesResponse> =>
    useQuery({
      queryKey: ["problemCategories"],
      queryFn: async () => await client.problemCategories({}, {}).response,
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
