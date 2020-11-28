import * as grpcWeb from 'grpc-web';

import * as library_checker_pb from './library_checker_pb';


export class LibraryCheckerServiceClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  register(
    request: library_checker_pb.RegisterRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.RegisterResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.RegisterResponse>;

  login(
    request: library_checker_pb.LoginRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.LoginResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.LoginResponse>;

  userInfo(
    request: library_checker_pb.UserInfoRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.UserInfoResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.UserInfoResponse>;

  userList(
    request: library_checker_pb.UserListRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.UserListResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.UserListResponse>;

  changeUserInfo(
    request: library_checker_pb.ChangeUserInfoRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.ChangeUserInfoResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.ChangeUserInfoResponse>;

  problemInfo(
    request: library_checker_pb.ProblemInfoRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.ProblemInfoResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.ProblemInfoResponse>;

  problemList(
    request: library_checker_pb.ProblemListRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.ProblemListResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.ProblemListResponse>;

  changeProblemInfo(
    request: library_checker_pb.ChangeProblemInfoRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.ChangeProblemInfoResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.ChangeProblemInfoResponse>;

  submit(
    request: library_checker_pb.SubmitRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.SubmitResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.SubmitResponse>;

  submissionInfo(
    request: library_checker_pb.SubmissionInfoRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.SubmissionInfoResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.SubmissionInfoResponse>;

  submissionList(
    request: library_checker_pb.SubmissionListRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.SubmissionListResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.SubmissionListResponse>;

  rejudge(
    request: library_checker_pb.RejudgeRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.RejudgeResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.RejudgeResponse>;

  langList(
    request: library_checker_pb.LangListRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.LangListResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.LangListResponse>;

  ranking(
    request: library_checker_pb.RankingRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.RankingResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.RankingResponse>;

  popJudgeTask(
    request: library_checker_pb.PopJudgeTaskRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.PopJudgeTaskResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.PopJudgeTaskResponse>;

  syncJudgeTaskStatus(
    request: library_checker_pb.SyncJudgeTaskStatusRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.SyncJudgeTaskStatusResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.SyncJudgeTaskStatusResponse>;

  finishJudgeTask(
    request: library_checker_pb.FinishJudgeTaskRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.Error,
               response: library_checker_pb.FinishJudgeTaskResponse) => void
  ): grpcWeb.ClientReadableStream<library_checker_pb.FinishJudgeTaskResponse>;

}

export class LibraryCheckerServicePromiseClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  register(
    request: library_checker_pb.RegisterRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.RegisterResponse>;

  login(
    request: library_checker_pb.LoginRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.LoginResponse>;

  userInfo(
    request: library_checker_pb.UserInfoRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.UserInfoResponse>;

  userList(
    request: library_checker_pb.UserListRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.UserListResponse>;

  changeUserInfo(
    request: library_checker_pb.ChangeUserInfoRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.ChangeUserInfoResponse>;

  problemInfo(
    request: library_checker_pb.ProblemInfoRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.ProblemInfoResponse>;

  problemList(
    request: library_checker_pb.ProblemListRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.ProblemListResponse>;

  changeProblemInfo(
    request: library_checker_pb.ChangeProblemInfoRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.ChangeProblemInfoResponse>;

  submit(
    request: library_checker_pb.SubmitRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.SubmitResponse>;

  submissionInfo(
    request: library_checker_pb.SubmissionInfoRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.SubmissionInfoResponse>;

  submissionList(
    request: library_checker_pb.SubmissionListRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.SubmissionListResponse>;

  rejudge(
    request: library_checker_pb.RejudgeRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.RejudgeResponse>;

  langList(
    request: library_checker_pb.LangListRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.LangListResponse>;

  ranking(
    request: library_checker_pb.RankingRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.RankingResponse>;

  popJudgeTask(
    request: library_checker_pb.PopJudgeTaskRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.PopJudgeTaskResponse>;

  syncJudgeTaskStatus(
    request: library_checker_pb.SyncJudgeTaskStatusRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.SyncJudgeTaskStatusResponse>;

  finishJudgeTask(
    request: library_checker_pb.FinishJudgeTaskRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<library_checker_pb.FinishJudgeTaskResponse>;

}

