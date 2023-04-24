import {
  useQuery,
  UseQueryOptions,
  UseQueryResult,
} from "@tanstack/react-query";
import { AuthState } from "../contexts/AuthContext";
import { LibraryCheckerServiceClient } from "./library_checker.client";
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
  UserInfoResponse,
} from "./library_checker";

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
    UseQueryOptions<SubmissionInfoResponse, unknown, SubmissionInfoResponse, string[]>,
    "queryKey" | "queryFn"
  >
): UseQueryResult<SubmissionInfoResponse> =>
  useQuery(
    ["submissionInfo2", String(id)],
    async () =>
        await client.submissionInfo(
        {
          id: id
        },
        state ? authMetadata(state) : undefined
      ).response,
      options
  );
