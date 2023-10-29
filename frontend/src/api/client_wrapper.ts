import {
  useMutation,
  UseMutationOptions,
  useQuery,
  useQueryClient,
  UseQueryOptions,
  UseQueryResult,
} from "@tanstack/react-query";
import { AuthState } from "../contexts/AuthContext";
import { LibraryCheckerServiceClient } from "../proto/library_checker.client";
import { GrpcWebFetchTransport } from "@protobuf-ts/grpcweb-transport";
import { RpcOptions } from "@protobuf-ts/runtime-rpc";
import {
  LangListResponse,
  ProblemCategoriesResponse,
  ProblemInfoResponse,
  ProblemListResponse,
  RankingResponse,
  SubmissionInfoResponse,
  SubmissionListResponse,
  SubmitRequest,
  SubmitResponse,
  UserInfoResponse,
} from "../proto/library_checker";
import { useCurrentAuthUser, useIdToken } from "../auth/auth";

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

const currentUserKey = ["api", "currentUser"];
export const useRegister = () => {
  const bearer = useBearer();
  const queryClient = useQueryClient();
  return useMutation(
    async (name: string) =>
      await client.register(
        {
          name: name,
        },
        bearer.data ?? undefined
      ),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(currentUserKey);
      },
    }
  );
};

export const useCurrentUser = () => {
  const bearer = useBearer();
  return useQuery(
    ["api", "currentUser", bearer.data],
    async () =>
      await client.currentUserInfo({}, bearer.data ?? undefined).response
  );
};

export const authMetadata = (state: AuthState): RpcOptions | undefined => {
  if (!state.token) {
    return undefined;
  } else {
    return {
      meta: {
        authorization: "bearer " + state.token,
      },
    };
  }
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
  name: string
): UseQueryResult<ProblemInfoResponse> =>
  useQuery(
    ["problemInfo", name],
    async () => await client.problemInfo({ name: name }, {}).response
  );

export const useProblemList = (): UseQueryResult<ProblemListResponse> =>
  useQuery(
    ["problemList"],
    async () => await client.problemList({}, {}).response
  );

export const useProblemCategories =
  (): UseQueryResult<ProblemCategoriesResponse> =>
    useQuery(
      ["problemCategories"],
      async () => await client.problemCategories({}, {}).response
    );

export const useUserInfo = (
  name: string,
  options?: Omit<
    UseQueryOptions<UserInfoResponse, unknown, UserInfoResponse, string[]>,
    "queryKey" | "queryFn"
  >
): UseQueryResult<UserInfoResponse> =>
  useQuery(
    ["userInfo", name],
    async () => await client.userInfo({ name: name }, {}).response,
    options
  );

export const useSubmissionList = (
  problem: string,
  user: string,
  status: string,
  lang: string,
  order: string,
  skip: number,
  limit: number
): UseQueryResult<SubmissionListResponse> =>
  useQuery(
    ["submissionList", problem, user, status, lang, order, skip, limit],
    async () =>
      await client.submissionList(
        {
          problem: problem,
          user: user,
          status: status,
          lang: lang,
          order: order,
          skip: skip,
          limit: limit,
          hacked: false,
        },
        {}
      ).response
  );

export const useSubmissionInfo = (
  id: number,
  state?: AuthState,
  options?: Omit<
    UseQueryOptions<
      SubmissionInfoResponse,
      unknown,
      SubmissionInfoResponse,
      string[]
    >,
    "queryKey" | "queryFn"
  >
): UseQueryResult<SubmissionInfoResponse> =>
  useQuery(
    ["submissionInfo2", String(id)],
    async () =>
      await client.submissionInfo(
        {
          id: id,
        },
        state ? authMetadata(state) : undefined
      ).response,
    options
  );

export const useSubmitMutation = (
  options?: Omit<
    UseMutationOptions<SubmitResponse, unknown, SubmitRequest, unknown>,
    "mutationFn"
  >
) => {
  const bearer = useBearer();
  return useMutation(
    async (req: SubmitRequest) =>
      await client.submit(req, bearer.data ?? undefined).response,
    options
  );
};
