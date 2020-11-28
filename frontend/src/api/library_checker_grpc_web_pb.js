/**
 * @fileoverview gRPC-Web generated client stub for librarychecker
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var google_protobuf_duration_pb = require('google-protobuf/google/protobuf/duration_pb.js')
const proto = {};
proto.librarychecker = require('./library_checker_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.librarychecker.LibraryCheckerServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.librarychecker.LibraryCheckerServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.RegisterRequest,
 *   !proto.librarychecker.RegisterResponse>}
 */
const methodDescriptor_LibraryCheckerService_Register = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/Register',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.RegisterRequest,
  proto.librarychecker.RegisterResponse,
  /**
   * @param {!proto.librarychecker.RegisterRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.RegisterResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.RegisterRequest,
 *   !proto.librarychecker.RegisterResponse>}
 */
const methodInfo_LibraryCheckerService_Register = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.RegisterResponse,
  /**
   * @param {!proto.librarychecker.RegisterRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.RegisterResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.RegisterRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.RegisterResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.RegisterResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.register =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Register',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Register,
      callback);
};


/**
 * @param {!proto.librarychecker.RegisterRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.RegisterResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.register =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Register',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Register);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.LoginRequest,
 *   !proto.librarychecker.LoginResponse>}
 */
const methodDescriptor_LibraryCheckerService_Login = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/Login',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.LoginRequest,
  proto.librarychecker.LoginResponse,
  /**
   * @param {!proto.librarychecker.LoginRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.LoginResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.LoginRequest,
 *   !proto.librarychecker.LoginResponse>}
 */
const methodInfo_LibraryCheckerService_Login = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.LoginResponse,
  /**
   * @param {!proto.librarychecker.LoginRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.LoginResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.LoginRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.LoginResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.LoginResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.login =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Login',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Login,
      callback);
};


/**
 * @param {!proto.librarychecker.LoginRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.LoginResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.login =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Login',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Login);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.UserInfoRequest,
 *   !proto.librarychecker.UserInfoResponse>}
 */
const methodDescriptor_LibraryCheckerService_UserInfo = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/UserInfo',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.UserInfoRequest,
  proto.librarychecker.UserInfoResponse,
  /**
   * @param {!proto.librarychecker.UserInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.UserInfoResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.UserInfoRequest,
 *   !proto.librarychecker.UserInfoResponse>}
 */
const methodInfo_LibraryCheckerService_UserInfo = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.UserInfoResponse,
  /**
   * @param {!proto.librarychecker.UserInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.UserInfoResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.UserInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.UserInfoResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.UserInfoResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.userInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/UserInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_UserInfo,
      callback);
};


/**
 * @param {!proto.librarychecker.UserInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.UserInfoResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.userInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/UserInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_UserInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.UserListRequest,
 *   !proto.librarychecker.UserListResponse>}
 */
const methodDescriptor_LibraryCheckerService_UserList = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/UserList',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.UserListRequest,
  proto.librarychecker.UserListResponse,
  /**
   * @param {!proto.librarychecker.UserListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.UserListResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.UserListRequest,
 *   !proto.librarychecker.UserListResponse>}
 */
const methodInfo_LibraryCheckerService_UserList = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.UserListResponse,
  /**
   * @param {!proto.librarychecker.UserListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.UserListResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.UserListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.UserListResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.UserListResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.userList =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/UserList',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_UserList,
      callback);
};


/**
 * @param {!proto.librarychecker.UserListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.UserListResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.userList =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/UserList',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_UserList);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.ChangeUserInfoRequest,
 *   !proto.librarychecker.ChangeUserInfoResponse>}
 */
const methodDescriptor_LibraryCheckerService_ChangeUserInfo = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/ChangeUserInfo',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.ChangeUserInfoRequest,
  proto.librarychecker.ChangeUserInfoResponse,
  /**
   * @param {!proto.librarychecker.ChangeUserInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.ChangeUserInfoResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.ChangeUserInfoRequest,
 *   !proto.librarychecker.ChangeUserInfoResponse>}
 */
const methodInfo_LibraryCheckerService_ChangeUserInfo = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.ChangeUserInfoResponse,
  /**
   * @param {!proto.librarychecker.ChangeUserInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.ChangeUserInfoResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.ChangeUserInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.ChangeUserInfoResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.ChangeUserInfoResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.changeUserInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/ChangeUserInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_ChangeUserInfo,
      callback);
};


/**
 * @param {!proto.librarychecker.ChangeUserInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.ChangeUserInfoResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.changeUserInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/ChangeUserInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_ChangeUserInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.ProblemInfoRequest,
 *   !proto.librarychecker.ProblemInfoResponse>}
 */
const methodDescriptor_LibraryCheckerService_ProblemInfo = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/ProblemInfo',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.ProblemInfoRequest,
  proto.librarychecker.ProblemInfoResponse,
  /**
   * @param {!proto.librarychecker.ProblemInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.ProblemInfoResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.ProblemInfoRequest,
 *   !proto.librarychecker.ProblemInfoResponse>}
 */
const methodInfo_LibraryCheckerService_ProblemInfo = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.ProblemInfoResponse,
  /**
   * @param {!proto.librarychecker.ProblemInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.ProblemInfoResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.ProblemInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.ProblemInfoResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.ProblemInfoResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.problemInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/ProblemInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_ProblemInfo,
      callback);
};


/**
 * @param {!proto.librarychecker.ProblemInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.ProblemInfoResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.problemInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/ProblemInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_ProblemInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.ProblemListRequest,
 *   !proto.librarychecker.ProblemListResponse>}
 */
const methodDescriptor_LibraryCheckerService_ProblemList = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/ProblemList',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.ProblemListRequest,
  proto.librarychecker.ProblemListResponse,
  /**
   * @param {!proto.librarychecker.ProblemListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.ProblemListResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.ProblemListRequest,
 *   !proto.librarychecker.ProblemListResponse>}
 */
const methodInfo_LibraryCheckerService_ProblemList = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.ProblemListResponse,
  /**
   * @param {!proto.librarychecker.ProblemListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.ProblemListResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.ProblemListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.ProblemListResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.ProblemListResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.problemList =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/ProblemList',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_ProblemList,
      callback);
};


/**
 * @param {!proto.librarychecker.ProblemListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.ProblemListResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.problemList =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/ProblemList',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_ProblemList);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.ChangeProblemInfoRequest,
 *   !proto.librarychecker.ChangeProblemInfoResponse>}
 */
const methodDescriptor_LibraryCheckerService_ChangeProblemInfo = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/ChangeProblemInfo',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.ChangeProblemInfoRequest,
  proto.librarychecker.ChangeProblemInfoResponse,
  /**
   * @param {!proto.librarychecker.ChangeProblemInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.ChangeProblemInfoResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.ChangeProblemInfoRequest,
 *   !proto.librarychecker.ChangeProblemInfoResponse>}
 */
const methodInfo_LibraryCheckerService_ChangeProblemInfo = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.ChangeProblemInfoResponse,
  /**
   * @param {!proto.librarychecker.ChangeProblemInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.ChangeProblemInfoResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.ChangeProblemInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.ChangeProblemInfoResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.ChangeProblemInfoResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.changeProblemInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/ChangeProblemInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_ChangeProblemInfo,
      callback);
};


/**
 * @param {!proto.librarychecker.ChangeProblemInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.ChangeProblemInfoResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.changeProblemInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/ChangeProblemInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_ChangeProblemInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.SubmitRequest,
 *   !proto.librarychecker.SubmitResponse>}
 */
const methodDescriptor_LibraryCheckerService_Submit = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/Submit',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.SubmitRequest,
  proto.librarychecker.SubmitResponse,
  /**
   * @param {!proto.librarychecker.SubmitRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.SubmitResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.SubmitRequest,
 *   !proto.librarychecker.SubmitResponse>}
 */
const methodInfo_LibraryCheckerService_Submit = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.SubmitResponse,
  /**
   * @param {!proto.librarychecker.SubmitRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.SubmitResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.SubmitRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.SubmitResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.SubmitResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.submit =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Submit',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Submit,
      callback);
};


/**
 * @param {!proto.librarychecker.SubmitRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.SubmitResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.submit =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Submit',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Submit);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.SubmissionInfoRequest,
 *   !proto.librarychecker.SubmissionInfoResponse>}
 */
const methodDescriptor_LibraryCheckerService_SubmissionInfo = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/SubmissionInfo',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.SubmissionInfoRequest,
  proto.librarychecker.SubmissionInfoResponse,
  /**
   * @param {!proto.librarychecker.SubmissionInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.SubmissionInfoResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.SubmissionInfoRequest,
 *   !proto.librarychecker.SubmissionInfoResponse>}
 */
const methodInfo_LibraryCheckerService_SubmissionInfo = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.SubmissionInfoResponse,
  /**
   * @param {!proto.librarychecker.SubmissionInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.SubmissionInfoResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.SubmissionInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.SubmissionInfoResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.SubmissionInfoResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.submissionInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/SubmissionInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_SubmissionInfo,
      callback);
};


/**
 * @param {!proto.librarychecker.SubmissionInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.SubmissionInfoResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.submissionInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/SubmissionInfo',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_SubmissionInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.SubmissionListRequest,
 *   !proto.librarychecker.SubmissionListResponse>}
 */
const methodDescriptor_LibraryCheckerService_SubmissionList = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/SubmissionList',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.SubmissionListRequest,
  proto.librarychecker.SubmissionListResponse,
  /**
   * @param {!proto.librarychecker.SubmissionListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.SubmissionListResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.SubmissionListRequest,
 *   !proto.librarychecker.SubmissionListResponse>}
 */
const methodInfo_LibraryCheckerService_SubmissionList = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.SubmissionListResponse,
  /**
   * @param {!proto.librarychecker.SubmissionListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.SubmissionListResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.SubmissionListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.SubmissionListResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.SubmissionListResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.submissionList =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/SubmissionList',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_SubmissionList,
      callback);
};


/**
 * @param {!proto.librarychecker.SubmissionListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.SubmissionListResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.submissionList =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/SubmissionList',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_SubmissionList);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.RejudgeRequest,
 *   !proto.librarychecker.RejudgeResponse>}
 */
const methodDescriptor_LibraryCheckerService_Rejudge = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/Rejudge',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.RejudgeRequest,
  proto.librarychecker.RejudgeResponse,
  /**
   * @param {!proto.librarychecker.RejudgeRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.RejudgeResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.RejudgeRequest,
 *   !proto.librarychecker.RejudgeResponse>}
 */
const methodInfo_LibraryCheckerService_Rejudge = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.RejudgeResponse,
  /**
   * @param {!proto.librarychecker.RejudgeRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.RejudgeResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.RejudgeRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.RejudgeResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.RejudgeResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.rejudge =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Rejudge',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Rejudge,
      callback);
};


/**
 * @param {!proto.librarychecker.RejudgeRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.RejudgeResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.rejudge =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Rejudge',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Rejudge);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.LangListRequest,
 *   !proto.librarychecker.LangListResponse>}
 */
const methodDescriptor_LibraryCheckerService_LangList = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/LangList',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.LangListRequest,
  proto.librarychecker.LangListResponse,
  /**
   * @param {!proto.librarychecker.LangListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.LangListResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.LangListRequest,
 *   !proto.librarychecker.LangListResponse>}
 */
const methodInfo_LibraryCheckerService_LangList = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.LangListResponse,
  /**
   * @param {!proto.librarychecker.LangListRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.LangListResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.LangListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.LangListResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.LangListResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.langList =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/LangList',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_LangList,
      callback);
};


/**
 * @param {!proto.librarychecker.LangListRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.LangListResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.langList =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/LangList',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_LangList);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.RankingRequest,
 *   !proto.librarychecker.RankingResponse>}
 */
const methodDescriptor_LibraryCheckerService_Ranking = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/Ranking',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.RankingRequest,
  proto.librarychecker.RankingResponse,
  /**
   * @param {!proto.librarychecker.RankingRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.RankingResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.RankingRequest,
 *   !proto.librarychecker.RankingResponse>}
 */
const methodInfo_LibraryCheckerService_Ranking = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.RankingResponse,
  /**
   * @param {!proto.librarychecker.RankingRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.RankingResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.RankingRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.RankingResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.RankingResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.ranking =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Ranking',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Ranking,
      callback);
};


/**
 * @param {!proto.librarychecker.RankingRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.RankingResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.ranking =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/Ranking',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_Ranking);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.PopJudgeTaskRequest,
 *   !proto.librarychecker.PopJudgeTaskResponse>}
 */
const methodDescriptor_LibraryCheckerService_PopJudgeTask = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/PopJudgeTask',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.PopJudgeTaskRequest,
  proto.librarychecker.PopJudgeTaskResponse,
  /**
   * @param {!proto.librarychecker.PopJudgeTaskRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.PopJudgeTaskResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.PopJudgeTaskRequest,
 *   !proto.librarychecker.PopJudgeTaskResponse>}
 */
const methodInfo_LibraryCheckerService_PopJudgeTask = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.PopJudgeTaskResponse,
  /**
   * @param {!proto.librarychecker.PopJudgeTaskRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.PopJudgeTaskResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.PopJudgeTaskRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.PopJudgeTaskResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.PopJudgeTaskResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.popJudgeTask =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/PopJudgeTask',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_PopJudgeTask,
      callback);
};


/**
 * @param {!proto.librarychecker.PopJudgeTaskRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.PopJudgeTaskResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.popJudgeTask =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/PopJudgeTask',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_PopJudgeTask);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.SyncJudgeTaskStatusRequest,
 *   !proto.librarychecker.SyncJudgeTaskStatusResponse>}
 */
const methodDescriptor_LibraryCheckerService_SyncJudgeTaskStatus = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/SyncJudgeTaskStatus',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.SyncJudgeTaskStatusRequest,
  proto.librarychecker.SyncJudgeTaskStatusResponse,
  /**
   * @param {!proto.librarychecker.SyncJudgeTaskStatusRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.SyncJudgeTaskStatusResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.SyncJudgeTaskStatusRequest,
 *   !proto.librarychecker.SyncJudgeTaskStatusResponse>}
 */
const methodInfo_LibraryCheckerService_SyncJudgeTaskStatus = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.SyncJudgeTaskStatusResponse,
  /**
   * @param {!proto.librarychecker.SyncJudgeTaskStatusRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.SyncJudgeTaskStatusResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.SyncJudgeTaskStatusRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.SyncJudgeTaskStatusResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.SyncJudgeTaskStatusResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.syncJudgeTaskStatus =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/SyncJudgeTaskStatus',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_SyncJudgeTaskStatus,
      callback);
};


/**
 * @param {!proto.librarychecker.SyncJudgeTaskStatusRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.SyncJudgeTaskStatusResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.syncJudgeTaskStatus =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/SyncJudgeTaskStatus',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_SyncJudgeTaskStatus);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.librarychecker.FinishJudgeTaskRequest,
 *   !proto.librarychecker.FinishJudgeTaskResponse>}
 */
const methodDescriptor_LibraryCheckerService_FinishJudgeTask = new grpc.web.MethodDescriptor(
  '/librarychecker.LibraryCheckerService/FinishJudgeTask',
  grpc.web.MethodType.UNARY,
  proto.librarychecker.FinishJudgeTaskRequest,
  proto.librarychecker.FinishJudgeTaskResponse,
  /**
   * @param {!proto.librarychecker.FinishJudgeTaskRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.FinishJudgeTaskResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.librarychecker.FinishJudgeTaskRequest,
 *   !proto.librarychecker.FinishJudgeTaskResponse>}
 */
const methodInfo_LibraryCheckerService_FinishJudgeTask = new grpc.web.AbstractClientBase.MethodInfo(
  proto.librarychecker.FinishJudgeTaskResponse,
  /**
   * @param {!proto.librarychecker.FinishJudgeTaskRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.librarychecker.FinishJudgeTaskResponse.deserializeBinary
);


/**
 * @param {!proto.librarychecker.FinishJudgeTaskRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.librarychecker.FinishJudgeTaskResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.librarychecker.FinishJudgeTaskResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.librarychecker.LibraryCheckerServiceClient.prototype.finishJudgeTask =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/FinishJudgeTask',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_FinishJudgeTask,
      callback);
};


/**
 * @param {!proto.librarychecker.FinishJudgeTaskRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.librarychecker.FinishJudgeTaskResponse>}
 *     Promise that resolves to the response
 */
proto.librarychecker.LibraryCheckerServicePromiseClient.prototype.finishJudgeTask =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/librarychecker.LibraryCheckerService/FinishJudgeTask',
      request,
      metadata || {},
      methodDescriptor_LibraryCheckerService_FinishJudgeTask);
};


module.exports = proto.librarychecker;

