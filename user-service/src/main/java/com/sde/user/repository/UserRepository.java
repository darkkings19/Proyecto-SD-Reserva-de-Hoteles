package com.sde.user.repository;

import com.sde.user.entity.UserEntity;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

/**
 * Repositorio de Spring Data JPA para la entidad UserEntity.
 * Extender JpaRepository proporciona todos los métodos CRUD básicos.
 */
@Repository
public interface UserRepository extends JpaRepository<UserEntity, UUID> {

    /**
     * Busca un usuario por su email.
     * Útil para el login y para recuperar el perfil si el sistema se maneja por emails.
     * @param email el email a buscar
     * @return un Optional con la entidad encontrada, o vacío
     */
    Optional<UserEntity> findByEmail(String email);

    /**
     * Verifica eficientemente si un usuario con un email ya existe.
     * Fundamental para las validaciones previas al registro.
     * @param email el email a verificar
     * @return true si existe, false si no
     */
    boolean existsByEmail(String email);
}
