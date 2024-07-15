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
  LangListResponse,
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
  return useMutation(
    async (name: string) =>
      await client.register(
        {
          name: name,
        },
        bearer.data ?? undefined,
      ),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(currentUserKey);
      },
    },
  );
};
export const useChangeCurrentUserInfoMutation = () => {
  const bearer = useBearer();
  const queryClient = useQueryClient();
  return useMutation(
    async (req: ChangeCurrentUserInfoRequest) =>
      await client.changeCurrentUserInfo(req, bearer.data ?? undefined)
        .response,
    {
      onSuccess: () => {
        queryClient.invalidateQueries(currentUserKey);
      },
    },
  );
};

const transport = new GrpcWebFetchTransport({
  baseUrl: import.meta.env.VITE_API_URL,
});
const client = new LibraryCheckerServiceClient(transport);
export default client;

export const useLangList = (): UseQueryResult<LangListResponse> =>
  useQuery(["langList"], async () => await client.langList({}, {}).response);

export const useRanking = (): UseQueryResult<RankingResponse> =>
  useQuery(["ranking"], async () => await client.ranking({}, {}).response);

export const useProblemInfo = (
  name: string,
): UseQueryResult<ProblemInfoResponse> =>
  useQuery(
    ["problemInfo", name],
    async () => await client.problemInfo({ name: name }, {}).response,
  );

export const useProblemList = (): UseQueryResult<ProblemListResponse> =>
  useQuery(
    ["problemList"],
    async () => await client.problemList({}, {}).response,
  );

export const useProblemCategories =
  (): UseQueryResult<ProblemCategoriesResponse> =>
    useQuery(
      ["problemCategories"],
      async () => await client.problemCategories({}, {}).response,
    );

export const useUserInfo = (
  name: string,
  options?: Omit<
    UseQueryOptions<UserInfoResponse, unknown, UserInfoResponse, string[]>,
    "queryKey" | "queryFn"
  >,
): UseQueryResult<UserInfoResponse> => {
  const bearer = useBearer();
  return useQuery(
    ["api", "userInfo", name, bearer.data?.meta.authorization ?? ""],
    async () =>
      await client.userInfo({ name: name }, bearer.data ?? undefined).response,
    options,
  );
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
  useQuery(
    [
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
    async () =>
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
  );

export const useSubmissionInfo = (
  id: number,
  options?: Omit<
    UseQueryOptions<
      SubmissionInfoResponse,
      unknown,
      SubmissionInfoResponse,
      string[]
    >,
    "queryKey" | "queryFn"
  >,
): UseQueryResult<SubmissionInfoResponse> => {
  const bearer = useBearer();
  return useQuery(
    ["submissionInfo", String(id)],
    async () =>
      await client.submissionInfo(
        {
          id: id,
        },
        bearer.data ?? undefined,
      ).response,
    options,
  );
};

export const useSubmitMutation = (
  options?: Omit<
    UseMutationOptions<SubmitResponse, unknown, SubmitRequest, unknown>,
    "mutationFn"
  >,
) => {
  const bearer = useBearer();
  return useMutation(
    async (req: SubmitRequest) =>
      await client.submit(req, bearer.data ?? undefined).response,
    options,
  );
};

export const useRejudgeMutation = () => {
  const bearer = useBearer();
  return useMutation(
    async (req: RejudgeRequest) =>
      await client.rejudge(req, bearer.data ?? undefined).response,
  );
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
