from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetUserRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class User(_message.Message):
    __slots__ = ("id", "nombre", "email", "rol", "telefono", "created_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    NOMBRE_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    ROL_FIELD_NUMBER: _ClassVar[int]
    TELEFONO_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    nombre: str
    email: str
    rol: int
    telefono: str
    created_at: str
    def __init__(self, id: _Optional[str] = ..., nombre: _Optional[str] = ..., email: _Optional[str] = ..., rol: _Optional[int] = ..., telefono: _Optional[str] = ..., created_at: _Optional[str] = ...) -> None: ...

class UserResponse(_message.Message):
    __slots__ = ("user",)
    USER_FIELD_NUMBER: _ClassVar[int]
    user: User
    def __init__(self, user: _Optional[_Union[User, _Mapping]] = ...) -> None: ...
