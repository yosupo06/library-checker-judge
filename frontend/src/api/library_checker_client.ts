import { useQuery, UseQueryOptions, UseQueryResult } from "react-query";
import { AuthState } from "../contexts/AuthContext";
import { LibraryCheckerServiceClient } from "./Library_checkerServiceClientPb";
import {
  LangListRequest,
  LangListResponse,
  ProblemCategoriesRequest,
  ProblemCategoriesResponse,
  ProblemInfoRequest,
  ProblemInfoResponse,
  ProblemListRequest,
  ProblemListResponse,
  RankingRequest,
  RankingResponse,
  SubmissionListRequest,
  SubmissionListResponse,
  UserInfoRequest,
  UserInfoResponse,
} from "./library_checker_pb";

const api_url = process.env.REACT_APP_API_URL;

export const authMetadata = (
  state: AuthState
):
  | {
      authorization: string;
    }
  | undefined => {
  if (!state.token) {
    return undefined;
  } else {
    return {
      authorization: "bearer " + state.token,
    };
  }
};

const client = new LibraryCheckerServiceClient(
  api_url ?? "https://grpcweb-apiv1.yosupo.jp:443"
);

export default client;

export const useLangList = (): UseQueryResult<LangListResponse> =>
  useQuery("langList", () => client.langList(new LangListRequest(), {}));

export const useRanking = (): UseQueryResult<RankingResponse> =>
  useQuery("ranking", () => client.ranking(new RankingRequest(), {}));

export const useProblemInfo = (
  name: string
): UseQueryResult<ProblemInfoResponse> =>
  useQuery(["problemInfo", name], () =>
    client.problemInfo(new ProblemInfoRequest().setName(name), {})
  );

export const useProblemList = (): UseQueryResult<ProblemListResponse> =>
  useQuery("problemList", () =>
    client.problemList(new ProblemListRequest(), {})
  );

export const useProblemCategories =
  (): UseQueryResult<ProblemCategoriesResponse> =>
    useQuery("problemCategories", () =>
      client.problemCategories(new ProblemCategoriesRequest(), {})
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
    () => client.userInfo(new UserInfoRequest().setName(name), {}),
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
    () =>
      client.submissionList(
        new SubmissionListRequest()
          .setProblem(problem)
          .setUser(user)
          .setStatus(status)
          .setLang(lang)
          .setOrder(order)
          .setSkip(skip)
          .setLimit(limit),
        {}
      )
  );
