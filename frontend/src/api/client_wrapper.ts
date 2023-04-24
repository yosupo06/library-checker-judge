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
  SubmissionListResponse,
  UserInfoResponse,
} from "./library_checker";

export const authMetadata = (
  state: AuthState
): RpcOptions | undefined => {
  if (!state.token) {
    return undefined;
  } else {
    return {
      meta: {
        authorization: "bearer " + state.token,
      }
    };
  }
};

const api_url = process.env.REACT_APP_API_URL;
const transport = new GrpcWebFetchTransport({
  baseUrl: api_url ?? "https://grpcweb-apiv1.yosupo.jp:443",
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
