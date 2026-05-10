package com.sde.user.grpc;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * Definición del contrato del microservicio de Usuarios
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.62.2)",
    comments = "Source: user.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class UserServiceGrpc {

  private UserServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "com.sde.user.grpc.UserService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.sde.user.grpc.CreateUserRequest,
      com.sde.user.grpc.UserResponse> getCreateUserMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateUser",
      requestType = com.sde.user.grpc.CreateUserRequest.class,
      responseType = com.sde.user.grpc.UserResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.sde.user.grpc.CreateUserRequest,
      com.sde.user.grpc.UserResponse> getCreateUserMethod() {
    io.grpc.MethodDescriptor<com.sde.user.grpc.CreateUserRequest, com.sde.user.grpc.UserResponse> getCreateUserMethod;
    if ((getCreateUserMethod = UserServiceGrpc.getCreateUserMethod) == null) {
      synchronized (UserServiceGrpc.class) {
        if ((getCreateUserMethod = UserServiceGrpc.getCreateUserMethod) == null) {
          UserServiceGrpc.getCreateUserMethod = getCreateUserMethod =
              io.grpc.MethodDescriptor.<com.sde.user.grpc.CreateUserRequest, com.sde.user.grpc.UserResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateUser"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.sde.user.grpc.CreateUserRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.sde.user.grpc.UserResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserServiceMethodDescriptorSupplier("CreateUser"))
              .build();
        }
      }
    }
    return getCreateUserMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.sde.user.grpc.GetUserRequest,
      com.sde.user.grpc.UserResponse> getGetUserMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetUser",
      requestType = com.sde.user.grpc.GetUserRequest.class,
      responseType = com.sde.user.grpc.UserResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.sde.user.grpc.GetUserRequest,
      com.sde.user.grpc.UserResponse> getGetUserMethod() {
    io.grpc.MethodDescriptor<com.sde.user.grpc.GetUserRequest, com.sde.user.grpc.UserResponse> getGetUserMethod;
    if ((getGetUserMethod = UserServiceGrpc.getGetUserMethod) == null) {
      synchronized (UserServiceGrpc.class) {
        if ((getGetUserMethod = UserServiceGrpc.getGetUserMethod) == null) {
          UserServiceGrpc.getGetUserMethod = getGetUserMethod =
              io.grpc.MethodDescriptor.<com.sde.user.grpc.GetUserRequest, com.sde.user.grpc.UserResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetUser"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.sde.user.grpc.GetUserRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.sde.user.grpc.UserResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserServiceMethodDescriptorSupplier("GetUser"))
              .build();
        }
      }
    }
    return getGetUserMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.sde.user.grpc.UpdateUserRequest,
      com.sde.user.grpc.UserResponse> getUpdateUserMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateUser",
      requestType = com.sde.user.grpc.UpdateUserRequest.class,
      responseType = com.sde.user.grpc.UserResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.sde.user.grpc.UpdateUserRequest,
      com.sde.user.grpc.UserResponse> getUpdateUserMethod() {
    io.grpc.MethodDescriptor<com.sde.user.grpc.UpdateUserRequest, com.sde.user.grpc.UserResponse> getUpdateUserMethod;
    if ((getUpdateUserMethod = UserServiceGrpc.getUpdateUserMethod) == null) {
      synchronized (UserServiceGrpc.class) {
        if ((getUpdateUserMethod = UserServiceGrpc.getUpdateUserMethod) == null) {
          UserServiceGrpc.getUpdateUserMethod = getUpdateUserMethod =
              io.grpc.MethodDescriptor.<com.sde.user.grpc.UpdateUserRequest, com.sde.user.grpc.UserResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateUser"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.sde.user.grpc.UpdateUserRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.sde.user.grpc.UserResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserServiceMethodDescriptorSupplier("UpdateUser"))
              .build();
        }
      }
    }
    return getUpdateUserMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.sde.user.grpc.AuthenticateUserRequest,
      com.sde.user.grpc.AuthenticateUserResponse> getAuthenticateUserMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "AuthenticateUser",
      requestType = com.sde.user.grpc.AuthenticateUserRequest.class,
      responseType = com.sde.user.grpc.AuthenticateUserResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.sde.user.grpc.AuthenticateUserRequest,
      com.sde.user.grpc.AuthenticateUserResponse> getAuthenticateUserMethod() {
    io.grpc.MethodDescriptor<com.sde.user.grpc.AuthenticateUserRequest, com.sde.user.grpc.AuthenticateUserResponse> getAuthenticateUserMethod;
    if ((getAuthenticateUserMethod = UserServiceGrpc.getAuthenticateUserMethod) == null) {
      synchronized (UserServiceGrpc.class) {
        if ((getAuthenticateUserMethod = UserServiceGrpc.getAuthenticateUserMethod) == null) {
          UserServiceGrpc.getAuthenticateUserMethod = getAuthenticateUserMethod =
              io.grpc.MethodDescriptor.<com.sde.user.grpc.AuthenticateUserRequest, com.sde.user.grpc.AuthenticateUserResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "AuthenticateUser"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.sde.user.grpc.AuthenticateUserRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.sde.user.grpc.AuthenticateUserResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserServiceMethodDescriptorSupplier("AuthenticateUser"))
              .build();
        }
      }
    }
    return getAuthenticateUserMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static UserServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserServiceStub>() {
        @java.lang.Override
        public UserServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserServiceStub(channel, callOptions);
        }
      };
    return UserServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static UserServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserServiceBlockingStub>() {
        @java.lang.Override
        public UserServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserServiceBlockingStub(channel, callOptions);
        }
      };
    return UserServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static UserServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserServiceFutureStub>() {
        @java.lang.Override
        public UserServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserServiceFutureStub(channel, callOptions);
        }
      };
    return UserServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * Definición del contrato del microservicio de Usuarios
   * </pre>
   */
  public interface AsyncService {

    /**
     * <pre>
     * Caso de Uso: Crear un nuevo usuario
     * </pre>
     */
    default void createUser(com.sde.user.grpc.CreateUserRequest request,
        io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateUserMethod(), responseObserver);
    }

    /**
     * <pre>
     * Caso de Uso: Obtener el perfil del usuario (ej. para Mi Perfil)
     * </pre>
     */
    default void getUser(com.sde.user.grpc.GetUserRequest request,
        io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetUserMethod(), responseObserver);
    }

    /**
     * <pre>
     * Caso de Uso: Actualizar los datos del usuario
     * </pre>
     */
    default void updateUser(com.sde.user.grpc.UpdateUserRequest request,
        io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateUserMethod(), responseObserver);
    }

    /**
     * <pre>
     * Caso de Uso: Autenticar a un usuario (Validar credenciales)
     * </pre>
     */
    default void authenticateUser(com.sde.user.grpc.AuthenticateUserRequest request,
        io.grpc.stub.StreamObserver<com.sde.user.grpc.AuthenticateUserResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getAuthenticateUserMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service UserService.
   * <pre>
   * Definición del contrato del microservicio de Usuarios
   * </pre>
   */
  public static abstract class UserServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return UserServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service UserService.
   * <pre>
   * Definición del contrato del microservicio de Usuarios
   * </pre>
   */
  public static final class UserServiceStub
      extends io.grpc.stub.AbstractAsyncStub<UserServiceStub> {
    private UserServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * Caso de Uso: Crear un nuevo usuario
     * </pre>
     */
    public void createUser(com.sde.user.grpc.CreateUserRequest request,
        io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateUserMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Caso de Uso: Obtener el perfil del usuario (ej. para Mi Perfil)
     * </pre>
     */
    public void getUser(com.sde.user.grpc.GetUserRequest request,
        io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetUserMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Caso de Uso: Actualizar los datos del usuario
     * </pre>
     */
    public void updateUser(com.sde.user.grpc.UpdateUserRequest request,
        io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateUserMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Caso de Uso: Autenticar a un usuario (Validar credenciales)
     * </pre>
     */
    public void authenticateUser(com.sde.user.grpc.AuthenticateUserRequest request,
        io.grpc.stub.StreamObserver<com.sde.user.grpc.AuthenticateUserResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getAuthenticateUserMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service UserService.
   * <pre>
   * Definición del contrato del microservicio de Usuarios
   * </pre>
   */
  public static final class UserServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<UserServiceBlockingStub> {
    private UserServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Caso de Uso: Crear un nuevo usuario
     * </pre>
     */
    public com.sde.user.grpc.UserResponse createUser(com.sde.user.grpc.CreateUserRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateUserMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Caso de Uso: Obtener el perfil del usuario (ej. para Mi Perfil)
     * </pre>
     */
    public com.sde.user.grpc.UserResponse getUser(com.sde.user.grpc.GetUserRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetUserMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Caso de Uso: Actualizar los datos del usuario
     * </pre>
     */
    public com.sde.user.grpc.UserResponse updateUser(com.sde.user.grpc.UpdateUserRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateUserMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Caso de Uso: Autenticar a un usuario (Validar credenciales)
     * </pre>
     */
    public com.sde.user.grpc.AuthenticateUserResponse authenticateUser(com.sde.user.grpc.AuthenticateUserRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getAuthenticateUserMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service UserService.
   * <pre>
   * Definición del contrato del microservicio de Usuarios
   * </pre>
   */
  public static final class UserServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<UserServiceFutureStub> {
    private UserServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Caso de Uso: Crear un nuevo usuario
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<com.sde.user.grpc.UserResponse> createUser(
        com.sde.user.grpc.CreateUserRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateUserMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Caso de Uso: Obtener el perfil del usuario (ej. para Mi Perfil)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<com.sde.user.grpc.UserResponse> getUser(
        com.sde.user.grpc.GetUserRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetUserMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Caso de Uso: Actualizar los datos del usuario
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<com.sde.user.grpc.UserResponse> updateUser(
        com.sde.user.grpc.UpdateUserRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateUserMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Caso de Uso: Autenticar a un usuario (Validar credenciales)
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<com.sde.user.grpc.AuthenticateUserResponse> authenticateUser(
        com.sde.user.grpc.AuthenticateUserRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getAuthenticateUserMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CREATE_USER = 0;
  private static final int METHODID_GET_USER = 1;
  private static final int METHODID_UPDATE_USER = 2;
  private static final int METHODID_AUTHENTICATE_USER = 3;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_CREATE_USER:
          serviceImpl.createUser((com.sde.user.grpc.CreateUserRequest) request,
              (io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse>) responseObserver);
          break;
        case METHODID_GET_USER:
          serviceImpl.getUser((com.sde.user.grpc.GetUserRequest) request,
              (io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse>) responseObserver);
          break;
        case METHODID_UPDATE_USER:
          serviceImpl.updateUser((com.sde.user.grpc.UpdateUserRequest) request,
              (io.grpc.stub.StreamObserver<com.sde.user.grpc.UserResponse>) responseObserver);
          break;
        case METHODID_AUTHENTICATE_USER:
          serviceImpl.authenticateUser((com.sde.user.grpc.AuthenticateUserRequest) request,
              (io.grpc.stub.StreamObserver<com.sde.user.grpc.AuthenticateUserResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getCreateUserMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.sde.user.grpc.CreateUserRequest,
              com.sde.user.grpc.UserResponse>(
                service, METHODID_CREATE_USER)))
        .addMethod(
          getGetUserMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.sde.user.grpc.GetUserRequest,
              com.sde.user.grpc.UserResponse>(
                service, METHODID_GET_USER)))
        .addMethod(
          getUpdateUserMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.sde.user.grpc.UpdateUserRequest,
              com.sde.user.grpc.UserResponse>(
                service, METHODID_UPDATE_USER)))
        .addMethod(
          getAuthenticateUserMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.sde.user.grpc.AuthenticateUserRequest,
              com.sde.user.grpc.AuthenticateUserResponse>(
                service, METHODID_AUTHENTICATE_USER)))
        .build();
  }

  private static abstract class UserServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UserServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.sde.user.grpc.UserProto.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("UserService");
    }
  }

  private static final class UserServiceFileDescriptorSupplier
      extends UserServiceBaseDescriptorSupplier {
    UserServiceFileDescriptorSupplier() {}
  }

  private static final class UserServiceMethodDescriptorSupplier
      extends UserServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    UserServiceMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (UserServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new UserServiceFileDescriptorSupplier())
              .addMethod(getCreateUserMethod())
              .addMethod(getGetUserMethod())
              .addMethod(getUpdateUserMethod())
              .addMethod(getAuthenticateUserMethod())
              .build();
        }
      }
    }
    return result;
  }
}
