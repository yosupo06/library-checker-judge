import * as jspb from 'google-protobuf'

import * as google_protobuf_duration_pb from 'google-protobuf/google/protobuf/duration_pb';


export class RegisterRequest extends jspb.Message {
  getName(): string;
  setName(value: string): RegisterRequest;

  getPassword(): string;
  setPassword(value: string): RegisterRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterRequest): RegisterRequest.AsObject;
  static serializeBinaryToWriter(message: RegisterRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterRequest;
  static deserializeBinaryFromReader(message: RegisterRequest, reader: jspb.BinaryReader): RegisterRequest;
}

export namespace RegisterRequest {
  export type AsObject = {
    name: string,
    password: string,
  }
}

export class RegisterResponse extends jspb.Message {
  getToken(): string;
  setToken(value: string): RegisterResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterResponse): RegisterResponse.AsObject;
  static serializeBinaryToWriter(message: RegisterResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterResponse;
  static deserializeBinaryFromReader(message: RegisterResponse, reader: jspb.BinaryReader): RegisterResponse;
}

export namespace RegisterResponse {
  export type AsObject = {
    token: string,
  }
}

export class LoginRequest extends jspb.Message {
  getName(): string;
  setName(value: string): LoginRequest;

  getPassword(): string;
  setPassword(value: string): LoginRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoginRequest.AsObject;
  static toObject(includeInstance: boolean, msg: LoginRequest): LoginRequest.AsObject;
  static serializeBinaryToWriter(message: LoginRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoginRequest;
  static deserializeBinaryFromReader(message: LoginRequest, reader: jspb.BinaryReader): LoginRequest;
}

export namespace LoginRequest {
  export type AsObject = {
    name: string,
    password: string,
  }
}

export class LoginResponse extends jspb.Message {
  getToken(): string;
  setToken(value: string): LoginResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoginResponse.AsObject;
  static toObject(includeInstance: boolean, msg: LoginResponse): LoginResponse.AsObject;
  static serializeBinaryToWriter(message: LoginResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoginResponse;
  static deserializeBinaryFromReader(message: LoginResponse, reader: jspb.BinaryReader): LoginResponse;
}

export namespace LoginResponse {
  export type AsObject = {
    token: string,
  }
}

export class User extends jspb.Message {
  getName(): string;
  setName(value: string): User;

  getIsAdmin(): boolean;
  setIsAdmin(value: boolean): User;

  getEmail(): string;
  setEmail(value: string): User;

  getLibraryUrl(): string;
  setLibraryUrl(value: string): User;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): User.AsObject;
  static toObject(includeInstance: boolean, msg: User): User.AsObject;
  static serializeBinaryToWriter(message: User, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): User;
  static deserializeBinaryFromReader(message: User, reader: jspb.BinaryReader): User;
}

export namespace User {
  export type AsObject = {
    name: string,
    isAdmin: boolean,
    email: string,
    libraryUrl: string,
  }
}

export class UserInfoRequest extends jspb.Message {
  getName(): string;
  setName(value: string): UserInfoRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UserInfoRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UserInfoRequest): UserInfoRequest.AsObject;
  static serializeBinaryToWriter(message: UserInfoRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UserInfoRequest;
  static deserializeBinaryFromReader(message: UserInfoRequest, reader: jspb.BinaryReader): UserInfoRequest;
}

export namespace UserInfoRequest {
  export type AsObject = {
    name: string,
  }
}

export class UserInfoResponse extends jspb.Message {
  getIsAdmin(): boolean;
  setIsAdmin(value: boolean): UserInfoResponse;

  getUser(): User | undefined;
  setUser(value?: User): UserInfoResponse;
  hasUser(): boolean;
  clearUser(): UserInfoResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UserInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: UserInfoResponse): UserInfoResponse.AsObject;
  static serializeBinaryToWriter(message: UserInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UserInfoResponse;
  static deserializeBinaryFromReader(message: UserInfoResponse, reader: jspb.BinaryReader): UserInfoResponse;
}

export namespace UserInfoResponse {
  export type AsObject = {
    isAdmin: boolean,
    user?: User.AsObject,
  }
}

export class UserListRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UserListRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UserListRequest): UserListRequest.AsObject;
  static serializeBinaryToWriter(message: UserListRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UserListRequest;
  static deserializeBinaryFromReader(message: UserListRequest, reader: jspb.BinaryReader): UserListRequest;
}

export namespace UserListRequest {
  export type AsObject = {
  }
}

export class UserListResponse extends jspb.Message {
  getUsersList(): Array<User>;
  setUsersList(value: Array<User>): UserListResponse;
  clearUsersList(): UserListResponse;
  addUsers(value?: User, index?: number): User;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UserListResponse.AsObject;
  static toObject(includeInstance: boolean, msg: UserListResponse): UserListResponse.AsObject;
  static serializeBinaryToWriter(message: UserListResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UserListResponse;
  static deserializeBinaryFromReader(message: UserListResponse, reader: jspb.BinaryReader): UserListResponse;
}

export namespace UserListResponse {
  export type AsObject = {
    usersList: Array<User.AsObject>,
  }
}

export class ChangeUserInfoRequest extends jspb.Message {
  getUser(): User | undefined;
  setUser(value?: User): ChangeUserInfoRequest;
  hasUser(): boolean;
  clearUser(): ChangeUserInfoRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChangeUserInfoRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ChangeUserInfoRequest): ChangeUserInfoRequest.AsObject;
  static serializeBinaryToWriter(message: ChangeUserInfoRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChangeUserInfoRequest;
  static deserializeBinaryFromReader(message: ChangeUserInfoRequest, reader: jspb.BinaryReader): ChangeUserInfoRequest;
}

export namespace ChangeUserInfoRequest {
  export type AsObject = {
    user?: User.AsObject,
  }
}

export class ChangeUserInfoResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChangeUserInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ChangeUserInfoResponse): ChangeUserInfoResponse.AsObject;
  static serializeBinaryToWriter(message: ChangeUserInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChangeUserInfoResponse;
  static deserializeBinaryFromReader(message: ChangeUserInfoResponse, reader: jspb.BinaryReader): ChangeUserInfoResponse;
}

export namespace ChangeUserInfoResponse {
  export type AsObject = {
  }
}

export class Problem extends jspb.Message {
  getName(): string;
  setName(value: string): Problem;

  getTitle(): string;
  setTitle(value: string): Problem;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Problem.AsObject;
  static toObject(includeInstance: boolean, msg: Problem): Problem.AsObject;
  static serializeBinaryToWriter(message: Problem, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Problem;
  static deserializeBinaryFromReader(message: Problem, reader: jspb.BinaryReader): Problem;
}

export namespace Problem {
  export type AsObject = {
    name: string,
    title: string,
  }
}

export class ProblemListRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ProblemListRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ProblemListRequest): ProblemListRequest.AsObject;
  static serializeBinaryToWriter(message: ProblemListRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ProblemListRequest;
  static deserializeBinaryFromReader(message: ProblemListRequest, reader: jspb.BinaryReader): ProblemListRequest;
}

export namespace ProblemListRequest {
  export type AsObject = {
  }
}

export class ProblemListResponse extends jspb.Message {
  getProblemsList(): Array<Problem>;
  setProblemsList(value: Array<Problem>): ProblemListResponse;
  clearProblemsList(): ProblemListResponse;
  addProblems(value?: Problem, index?: number): Problem;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ProblemListResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ProblemListResponse): ProblemListResponse.AsObject;
  static serializeBinaryToWriter(message: ProblemListResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ProblemListResponse;
  static deserializeBinaryFromReader(message: ProblemListResponse, reader: jspb.BinaryReader): ProblemListResponse;
}

export namespace ProblemListResponse {
  export type AsObject = {
    problemsList: Array<Problem.AsObject>,
  }
}

export class ProblemInfoRequest extends jspb.Message {
  getName(): string;
  setName(value: string): ProblemInfoRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ProblemInfoRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ProblemInfoRequest): ProblemInfoRequest.AsObject;
  static serializeBinaryToWriter(message: ProblemInfoRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ProblemInfoRequest;
  static deserializeBinaryFromReader(message: ProblemInfoRequest, reader: jspb.BinaryReader): ProblemInfoRequest;
}

export namespace ProblemInfoRequest {
  export type AsObject = {
    name: string,
  }
}

export class ProblemInfoResponse extends jspb.Message {
  getTitle(): string;
  setTitle(value: string): ProblemInfoResponse;

  getStatement(): string;
  setStatement(value: string): ProblemInfoResponse;

  getTimeLimit(): number;
  setTimeLimit(value: number): ProblemInfoResponse;

  getCaseVersion(): string;
  setCaseVersion(value: string): ProblemInfoResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ProblemInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ProblemInfoResponse): ProblemInfoResponse.AsObject;
  static serializeBinaryToWriter(message: ProblemInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ProblemInfoResponse;
  static deserializeBinaryFromReader(message: ProblemInfoResponse, reader: jspb.BinaryReader): ProblemInfoResponse;
}

export namespace ProblemInfoResponse {
  export type AsObject = {
    title: string,
    statement: string,
    timeLimit: number,
    caseVersion: string,
  }
}

export class ChangeProblemInfoRequest extends jspb.Message {
  getName(): string;
  setName(value: string): ChangeProblemInfoRequest;

  getTitle(): string;
  setTitle(value: string): ChangeProblemInfoRequest;

  getStatement(): string;
  setStatement(value: string): ChangeProblemInfoRequest;

  getTimeLimit(): number;
  setTimeLimit(value: number): ChangeProblemInfoRequest;

  getCaseVersion(): string;
  setCaseVersion(value: string): ChangeProblemInfoRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChangeProblemInfoRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ChangeProblemInfoRequest): ChangeProblemInfoRequest.AsObject;
  static serializeBinaryToWriter(message: ChangeProblemInfoRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChangeProblemInfoRequest;
  static deserializeBinaryFromReader(message: ChangeProblemInfoRequest, reader: jspb.BinaryReader): ChangeProblemInfoRequest;
}

export namespace ChangeProblemInfoRequest {
  export type AsObject = {
    name: string,
    title: string,
    statement: string,
    timeLimit: number,
    caseVersion: string,
  }
}

export class ChangeProblemInfoResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChangeProblemInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ChangeProblemInfoResponse): ChangeProblemInfoResponse.AsObject;
  static serializeBinaryToWriter(message: ChangeProblemInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChangeProblemInfoResponse;
  static deserializeBinaryFromReader(message: ChangeProblemInfoResponse, reader: jspb.BinaryReader): ChangeProblemInfoResponse;
}

export namespace ChangeProblemInfoResponse {
  export type AsObject = {
  }
}

export class SubmitRequest extends jspb.Message {
  getProblem(): string;
  setProblem(value: string): SubmitRequest;

  getSource(): string;
  setSource(value: string): SubmitRequest;

  getLang(): string;
  setLang(value: string): SubmitRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmitRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubmitRequest): SubmitRequest.AsObject;
  static serializeBinaryToWriter(message: SubmitRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmitRequest;
  static deserializeBinaryFromReader(message: SubmitRequest, reader: jspb.BinaryReader): SubmitRequest;
}

export namespace SubmitRequest {
  export type AsObject = {
    problem: string,
    source: string,
    lang: string,
  }
}

export class SubmitResponse extends jspb.Message {
  getId(): number;
  setId(value: number): SubmitResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmitResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SubmitResponse): SubmitResponse.AsObject;
  static serializeBinaryToWriter(message: SubmitResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmitResponse;
  static deserializeBinaryFromReader(message: SubmitResponse, reader: jspb.BinaryReader): SubmitResponse;
}

export namespace SubmitResponse {
  export type AsObject = {
    id: number,
  }
}

export class SubmissionOverview extends jspb.Message {
  getId(): number;
  setId(value: number): SubmissionOverview;

  getProblemName(): string;
  setProblemName(value: string): SubmissionOverview;

  getProblemTitle(): string;
  setProblemTitle(value: string): SubmissionOverview;

  getUserName(): string;
  setUserName(value: string): SubmissionOverview;

  getLang(): string;
  setLang(value: string): SubmissionOverview;

  getIsLatest(): boolean;
  setIsLatest(value: boolean): SubmissionOverview;

  getStatus(): string;
  setStatus(value: string): SubmissionOverview;

  getHacked(): boolean;
  setHacked(value: boolean): SubmissionOverview;

  getTime(): number;
  setTime(value: number): SubmissionOverview;

  getMemory(): number;
  setMemory(value: number): SubmissionOverview;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionOverview.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionOverview): SubmissionOverview.AsObject;
  static serializeBinaryToWriter(message: SubmissionOverview, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionOverview;
  static deserializeBinaryFromReader(message: SubmissionOverview, reader: jspb.BinaryReader): SubmissionOverview;
}

export namespace SubmissionOverview {
  export type AsObject = {
    id: number,
    problemName: string,
    problemTitle: string,
    userName: string,
    lang: string,
    isLatest: boolean,
    status: string,
    hacked: boolean,
    time: number,
    memory: number,
  }
}

export class SubmissionCaseResult extends jspb.Message {
  getCase(): string;
  setCase(value: string): SubmissionCaseResult;

  getStatus(): string;
  setStatus(value: string): SubmissionCaseResult;

  getTime(): number;
  setTime(value: number): SubmissionCaseResult;

  getMemory(): number;
  setMemory(value: number): SubmissionCaseResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionCaseResult.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionCaseResult): SubmissionCaseResult.AsObject;
  static serializeBinaryToWriter(message: SubmissionCaseResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionCaseResult;
  static deserializeBinaryFromReader(message: SubmissionCaseResult, reader: jspb.BinaryReader): SubmissionCaseResult;
}

export namespace SubmissionCaseResult {
  export type AsObject = {
    pb_case: string,
    status: string,
    time: number,
    memory: number,
  }
}

export class SubmissionInfoRequest extends jspb.Message {
  getId(): number;
  setId(value: number): SubmissionInfoRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionInfoRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionInfoRequest): SubmissionInfoRequest.AsObject;
  static serializeBinaryToWriter(message: SubmissionInfoRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionInfoRequest;
  static deserializeBinaryFromReader(message: SubmissionInfoRequest, reader: jspb.BinaryReader): SubmissionInfoRequest;
}

export namespace SubmissionInfoRequest {
  export type AsObject = {
    id: number,
  }
}

export class SubmissionInfoResponse extends jspb.Message {
  getOverview(): SubmissionOverview | undefined;
  setOverview(value?: SubmissionOverview): SubmissionInfoResponse;
  hasOverview(): boolean;
  clearOverview(): SubmissionInfoResponse;

  getCaseResultsList(): Array<SubmissionCaseResult>;
  setCaseResultsList(value: Array<SubmissionCaseResult>): SubmissionInfoResponse;
  clearCaseResultsList(): SubmissionInfoResponse;
  addCaseResults(value?: SubmissionCaseResult, index?: number): SubmissionCaseResult;

  getSource(): string;
  setSource(value: string): SubmissionInfoResponse;

  getCompileError(): Uint8Array | string;
  getCompileError_asU8(): Uint8Array;
  getCompileError_asB64(): string;
  setCompileError(value: Uint8Array | string): SubmissionInfoResponse;

  getCanRejudge(): boolean;
  setCanRejudge(value: boolean): SubmissionInfoResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionInfoResponse): SubmissionInfoResponse.AsObject;
  static serializeBinaryToWriter(message: SubmissionInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionInfoResponse;
  static deserializeBinaryFromReader(message: SubmissionInfoResponse, reader: jspb.BinaryReader): SubmissionInfoResponse;
}

export namespace SubmissionInfoResponse {
  export type AsObject = {
    overview?: SubmissionOverview.AsObject,
    caseResultsList: Array<SubmissionCaseResult.AsObject>,
    source: string,
    compileError: Uint8Array | string,
    canRejudge: boolean,
  }
}

export class SubmissionListRequest extends jspb.Message {
  getSkip(): number;
  setSkip(value: number): SubmissionListRequest;

  getLimit(): number;
  setLimit(value: number): SubmissionListRequest;

  getProblem(): string;
  setProblem(value: string): SubmissionListRequest;

  getStatus(): string;
  setStatus(value: string): SubmissionListRequest;

  getHacked(): boolean;
  setHacked(value: boolean): SubmissionListRequest;

  getUser(): string;
  setUser(value: string): SubmissionListRequest;

  getLang(): string;
  setLang(value: string): SubmissionListRequest;

  getOrder(): string;
  setOrder(value: string): SubmissionListRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionListRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionListRequest): SubmissionListRequest.AsObject;
  static serializeBinaryToWriter(message: SubmissionListRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionListRequest;
  static deserializeBinaryFromReader(message: SubmissionListRequest, reader: jspb.BinaryReader): SubmissionListRequest;
}

export namespace SubmissionListRequest {
  export type AsObject = {
    skip: number,
    limit: number,
    problem: string,
    status: string,
    hacked: boolean,
    user: string,
    lang: string,
    order: string,
  }
}

export class SubmissionListResponse extends jspb.Message {
  getSubmissionsList(): Array<SubmissionOverview>;
  setSubmissionsList(value: Array<SubmissionOverview>): SubmissionListResponse;
  clearSubmissionsList(): SubmissionListResponse;
  addSubmissions(value?: SubmissionOverview, index?: number): SubmissionOverview;

  getCount(): number;
  setCount(value: number): SubmissionListResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionListResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionListResponse): SubmissionListResponse.AsObject;
  static serializeBinaryToWriter(message: SubmissionListResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionListResponse;
  static deserializeBinaryFromReader(message: SubmissionListResponse, reader: jspb.BinaryReader): SubmissionListResponse;
}

export namespace SubmissionListResponse {
  export type AsObject = {
    submissionsList: Array<SubmissionOverview.AsObject>,
    count: number,
  }
}

export class RejudgeRequest extends jspb.Message {
  getId(): number;
  setId(value: number): RejudgeRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RejudgeRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RejudgeRequest): RejudgeRequest.AsObject;
  static serializeBinaryToWriter(message: RejudgeRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RejudgeRequest;
  static deserializeBinaryFromReader(message: RejudgeRequest, reader: jspb.BinaryReader): RejudgeRequest;
}

export namespace RejudgeRequest {
  export type AsObject = {
    id: number,
  }
}

export class RejudgeResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RejudgeResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RejudgeResponse): RejudgeResponse.AsObject;
  static serializeBinaryToWriter(message: RejudgeResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RejudgeResponse;
  static deserializeBinaryFromReader(message: RejudgeResponse, reader: jspb.BinaryReader): RejudgeResponse;
}

export namespace RejudgeResponse {
  export type AsObject = {
  }
}

export class Lang extends jspb.Message {
  getId(): string;
  setId(value: string): Lang;

  getName(): string;
  setName(value: string): Lang;

  getVersion(): string;
  setVersion(value: string): Lang;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Lang.AsObject;
  static toObject(includeInstance: boolean, msg: Lang): Lang.AsObject;
  static serializeBinaryToWriter(message: Lang, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Lang;
  static deserializeBinaryFromReader(message: Lang, reader: jspb.BinaryReader): Lang;
}

export namespace Lang {
  export type AsObject = {
    id: string,
    name: string,
    version: string,
  }
}

export class LangListRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LangListRequest.AsObject;
  static toObject(includeInstance: boolean, msg: LangListRequest): LangListRequest.AsObject;
  static serializeBinaryToWriter(message: LangListRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LangListRequest;
  static deserializeBinaryFromReader(message: LangListRequest, reader: jspb.BinaryReader): LangListRequest;
}

export namespace LangListRequest {
  export type AsObject = {
  }
}

export class LangListResponse extends jspb.Message {
  getLangsList(): Array<Lang>;
  setLangsList(value: Array<Lang>): LangListResponse;
  clearLangsList(): LangListResponse;
  addLangs(value?: Lang, index?: number): Lang;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LangListResponse.AsObject;
  static toObject(includeInstance: boolean, msg: LangListResponse): LangListResponse.AsObject;
  static serializeBinaryToWriter(message: LangListResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LangListResponse;
  static deserializeBinaryFromReader(message: LangListResponse, reader: jspb.BinaryReader): LangListResponse;
}

export namespace LangListResponse {
  export type AsObject = {
    langsList: Array<Lang.AsObject>,
  }
}

export class UserStatistics extends jspb.Message {
  getName(): string;
  setName(value: string): UserStatistics;

  getCount(): number;
  setCount(value: number): UserStatistics;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UserStatistics.AsObject;
  static toObject(includeInstance: boolean, msg: UserStatistics): UserStatistics.AsObject;
  static serializeBinaryToWriter(message: UserStatistics, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UserStatistics;
  static deserializeBinaryFromReader(message: UserStatistics, reader: jspb.BinaryReader): UserStatistics;
}

export namespace UserStatistics {
  export type AsObject = {
    name: string,
    count: number,
  }
}

export class RankingRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RankingRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RankingRequest): RankingRequest.AsObject;
  static serializeBinaryToWriter(message: RankingRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RankingRequest;
  static deserializeBinaryFromReader(message: RankingRequest, reader: jspb.BinaryReader): RankingRequest;
}

export namespace RankingRequest {
  export type AsObject = {
  }
}

export class RankingResponse extends jspb.Message {
  getStatisticsList(): Array<UserStatistics>;
  setStatisticsList(value: Array<UserStatistics>): RankingResponse;
  clearStatisticsList(): RankingResponse;
  addStatistics(value?: UserStatistics, index?: number): UserStatistics;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RankingResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RankingResponse): RankingResponse.AsObject;
  static serializeBinaryToWriter(message: RankingResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RankingResponse;
  static deserializeBinaryFromReader(message: RankingResponse, reader: jspb.BinaryReader): RankingResponse;
}

export namespace RankingResponse {
  export type AsObject = {
    statisticsList: Array<UserStatistics.AsObject>,
  }
}

export class PopJudgeTaskRequest extends jspb.Message {
  getJudgeName(): string;
  setJudgeName(value: string): PopJudgeTaskRequest;

  getExpectedTime(): google_protobuf_duration_pb.Duration | undefined;
  setExpectedTime(value?: google_protobuf_duration_pb.Duration): PopJudgeTaskRequest;
  hasExpectedTime(): boolean;
  clearExpectedTime(): PopJudgeTaskRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PopJudgeTaskRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PopJudgeTaskRequest): PopJudgeTaskRequest.AsObject;
  static serializeBinaryToWriter(message: PopJudgeTaskRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PopJudgeTaskRequest;
  static deserializeBinaryFromReader(message: PopJudgeTaskRequest, reader: jspb.BinaryReader): PopJudgeTaskRequest;
}

export namespace PopJudgeTaskRequest {
  export type AsObject = {
    judgeName: string,
    expectedTime?: google_protobuf_duration_pb.Duration.AsObject,
  }
}

export class PopJudgeTaskResponse extends jspb.Message {
  getSubmissionId(): number;
  setSubmissionId(value: number): PopJudgeTaskResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PopJudgeTaskResponse.AsObject;
  static toObject(includeInstance: boolean, msg: PopJudgeTaskResponse): PopJudgeTaskResponse.AsObject;
  static serializeBinaryToWriter(message: PopJudgeTaskResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PopJudgeTaskResponse;
  static deserializeBinaryFromReader(message: PopJudgeTaskResponse, reader: jspb.BinaryReader): PopJudgeTaskResponse;
}

export namespace PopJudgeTaskResponse {
  export type AsObject = {
    submissionId: number,
  }
}

export class SyncJudgeTaskStatusRequest extends jspb.Message {
  getJudgeName(): string;
  setJudgeName(value: string): SyncJudgeTaskStatusRequest;

  getSubmissionId(): number;
  setSubmissionId(value: number): SyncJudgeTaskStatusRequest;

  getStatus(): string;
  setStatus(value: string): SyncJudgeTaskStatusRequest;

  getTime(): number;
  setTime(value: number): SyncJudgeTaskStatusRequest;

  getMemory(): number;
  setMemory(value: number): SyncJudgeTaskStatusRequest;

  getCompileError(): Uint8Array | string;
  getCompileError_asU8(): Uint8Array;
  getCompileError_asB64(): string;
  setCompileError(value: Uint8Array | string): SyncJudgeTaskStatusRequest;

  getCaseResultsList(): Array<SubmissionCaseResult>;
  setCaseResultsList(value: Array<SubmissionCaseResult>): SyncJudgeTaskStatusRequest;
  clearCaseResultsList(): SyncJudgeTaskStatusRequest;
  addCaseResults(value?: SubmissionCaseResult, index?: number): SubmissionCaseResult;

  getExpectedTime(): google_protobuf_duration_pb.Duration | undefined;
  setExpectedTime(value?: google_protobuf_duration_pb.Duration): SyncJudgeTaskStatusRequest;
  hasExpectedTime(): boolean;
  clearExpectedTime(): SyncJudgeTaskStatusRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SyncJudgeTaskStatusRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SyncJudgeTaskStatusRequest): SyncJudgeTaskStatusRequest.AsObject;
  static serializeBinaryToWriter(message: SyncJudgeTaskStatusRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SyncJudgeTaskStatusRequest;
  static deserializeBinaryFromReader(message: SyncJudgeTaskStatusRequest, reader: jspb.BinaryReader): SyncJudgeTaskStatusRequest;
}

export namespace SyncJudgeTaskStatusRequest {
  export type AsObject = {
    judgeName: string,
    submissionId: number,
    status: string,
    time: number,
    memory: number,
    compileError: Uint8Array | string,
    caseResultsList: Array<SubmissionCaseResult.AsObject>,
    expectedTime?: google_protobuf_duration_pb.Duration.AsObject,
  }
}

export class SyncJudgeTaskStatusResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SyncJudgeTaskStatusResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SyncJudgeTaskStatusResponse): SyncJudgeTaskStatusResponse.AsObject;
  static serializeBinaryToWriter(message: SyncJudgeTaskStatusResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SyncJudgeTaskStatusResponse;
  static deserializeBinaryFromReader(message: SyncJudgeTaskStatusResponse, reader: jspb.BinaryReader): SyncJudgeTaskStatusResponse;
}

export namespace SyncJudgeTaskStatusResponse {
  export type AsObject = {
  }
}

export class FinishJudgeTaskRequest extends jspb.Message {
  getJudgeName(): string;
  setJudgeName(value: string): FinishJudgeTaskRequest;

  getSubmissionId(): number;
  setSubmissionId(value: number): FinishJudgeTaskRequest;

  getStatus(): string;
  setStatus(value: string): FinishJudgeTaskRequest;

  getTime(): number;
  setTime(value: number): FinishJudgeTaskRequest;

  getMemory(): number;
  setMemory(value: number): FinishJudgeTaskRequest;

  getCaseVersion(): string;
  setCaseVersion(value: string): FinishJudgeTaskRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FinishJudgeTaskRequest.AsObject;
  static toObject(includeInstance: boolean, msg: FinishJudgeTaskRequest): FinishJudgeTaskRequest.AsObject;
  static serializeBinaryToWriter(message: FinishJudgeTaskRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FinishJudgeTaskRequest;
  static deserializeBinaryFromReader(message: FinishJudgeTaskRequest, reader: jspb.BinaryReader): FinishJudgeTaskRequest;
}

export namespace FinishJudgeTaskRequest {
  export type AsObject = {
    judgeName: string,
    submissionId: number,
    status: string,
    time: number,
    memory: number,
    caseVersion: string,
  }
}

export class FinishJudgeTaskResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FinishJudgeTaskResponse.AsObject;
  static toObject(includeInstance: boolean, msg: FinishJudgeTaskResponse): FinishJudgeTaskResponse.AsObject;
  static serializeBinaryToWriter(message: FinishJudgeTaskResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FinishJudgeTaskResponse;
  static deserializeBinaryFromReader(message: FinishJudgeTaskResponse, reader: jspb.BinaryReader): FinishJudgeTaskResponse;
}

export namespace FinishJudgeTaskResponse {
  export type AsObject = {
  }
}

