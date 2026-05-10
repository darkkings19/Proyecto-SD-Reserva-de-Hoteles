package com.sde.user.service;

import com.sde.user.dto.CreateUserDto;
import com.sde.user.dto.UpdateUserDto;
import com.sde.user.dto.UserDto;

import java.util.UUID;

/**
 * Interfaz que define los casos de uso principales del dominio de usuarios.
 * Totalmente desacoplada de gRPC, HTTP o cualquier capa de transporte.
 */
public interface UserService {
    
    UserDto createUser(CreateUserDto createDto);
    
    UserDto getUserById(UUID id);
    
    UserDto updateUser(UUID id, UpdateUserDto updateDto);
    
    UserDto authenticate(String email, String password);
}
