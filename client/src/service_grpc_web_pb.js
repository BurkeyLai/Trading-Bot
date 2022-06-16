/**
 * @fileoverview gRPC-Web generated client stub for proto
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');

const proto = {};
proto.proto = require('./service_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.proto.TradingBotClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

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
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.proto.TradingBotPromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

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
 *   !proto.proto.Connect,
 *   !proto.proto.Message>}
 */
const methodDescriptor_TradingBot_CreateStream = new grpc.web.MethodDescriptor(
  '/proto.TradingBot/CreateStream',
  grpc.web.MethodType.SERVER_STREAMING,
  proto.proto.Connect,
  proto.proto.Message,
  /**
   * @param {!proto.proto.Connect} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.proto.Message.deserializeBinary
);


/**
 * @param {!proto.proto.Connect} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.proto.Message>}
 *     The XHR Node Readable Stream
 */
proto.proto.TradingBotClient.prototype.createStream =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/proto.TradingBot/CreateStream',
      request,
      metadata || {},
      methodDescriptor_TradingBot_CreateStream);
};


/**
 * @param {!proto.proto.Connect} request The request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.proto.Message>}
 *     The XHR Node Readable Stream
 */
proto.proto.TradingBotPromiseClient.prototype.createStream =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/proto.TradingBot/CreateStream',
      request,
      metadata || {},
      methodDescriptor_TradingBot_CreateStream);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.proto.MarketInfoRequest,
 *   !proto.proto.MarketInfoRespond>}
 */
const methodDescriptor_TradingBot_MarketInfo = new grpc.web.MethodDescriptor(
  '/proto.TradingBot/MarketInfo',
  grpc.web.MethodType.UNARY,
  proto.proto.MarketInfoRequest,
  proto.proto.MarketInfoRespond,
  /**
   * @param {!proto.proto.MarketInfoRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.proto.MarketInfoRespond.deserializeBinary
);


/**
 * @param {!proto.proto.MarketInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.proto.MarketInfoRespond)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.proto.MarketInfoRespond>|undefined}
 *     The XHR Node Readable Stream
 */
proto.proto.TradingBotClient.prototype.marketInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/proto.TradingBot/MarketInfo',
      request,
      metadata || {},
      methodDescriptor_TradingBot_MarketInfo,
      callback);
};


/**
 * @param {!proto.proto.MarketInfoRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.proto.MarketInfoRespond>}
 *     Promise that resolves to the response
 */
proto.proto.TradingBotPromiseClient.prototype.marketInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/proto.TradingBot/MarketInfo',
      request,
      metadata || {},
      methodDescriptor_TradingBot_MarketInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.proto.CreateOrderRequest,
 *   !proto.proto.CreateOrderRespond>}
 */
const methodDescriptor_TradingBot_CreateOrder = new grpc.web.MethodDescriptor(
  '/proto.TradingBot/CreateOrder',
  grpc.web.MethodType.UNARY,
  proto.proto.CreateOrderRequest,
  proto.proto.CreateOrderRespond,
  /**
   * @param {!proto.proto.CreateOrderRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.proto.CreateOrderRespond.deserializeBinary
);


/**
 * @param {!proto.proto.CreateOrderRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.proto.CreateOrderRespond)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.proto.CreateOrderRespond>|undefined}
 *     The XHR Node Readable Stream
 */
proto.proto.TradingBotClient.prototype.createOrder =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/proto.TradingBot/CreateOrder',
      request,
      metadata || {},
      methodDescriptor_TradingBot_CreateOrder,
      callback);
};


/**
 * @param {!proto.proto.CreateOrderRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.proto.CreateOrderRespond>}
 *     Promise that resolves to the response
 */
proto.proto.TradingBotPromiseClient.prototype.createOrder =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/proto.TradingBot/CreateOrder',
      request,
      metadata || {},
      methodDescriptor_TradingBot_CreateOrder);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.proto.AccountBalanceRequest,
 *   !proto.proto.AccountBalanceRespond>}
 */
const methodDescriptor_TradingBot_AccountBalance = new grpc.web.MethodDescriptor(
  '/proto.TradingBot/AccountBalance',
  grpc.web.MethodType.UNARY,
  proto.proto.AccountBalanceRequest,
  proto.proto.AccountBalanceRespond,
  /**
   * @param {!proto.proto.AccountBalanceRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.proto.AccountBalanceRespond.deserializeBinary
);


/**
 * @param {!proto.proto.AccountBalanceRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.proto.AccountBalanceRespond)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.proto.AccountBalanceRespond>|undefined}
 *     The XHR Node Readable Stream
 */
proto.proto.TradingBotClient.prototype.accountBalance =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/proto.TradingBot/AccountBalance',
      request,
      metadata || {},
      methodDescriptor_TradingBot_AccountBalance,
      callback);
};


/**
 * @param {!proto.proto.AccountBalanceRequest} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.proto.AccountBalanceRespond>}
 *     Promise that resolves to the response
 */
proto.proto.TradingBotPromiseClient.prototype.accountBalance =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/proto.TradingBot/AccountBalance',
      request,
      metadata || {},
      methodDescriptor_TradingBot_AccountBalance);
};


module.exports = proto.proto;

