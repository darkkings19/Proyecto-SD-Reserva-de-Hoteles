package com.sde.user.entity;

import jakarta.persistence.*;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import org.hibernate.annotations.CreationTimestamp;

import java.time.Instant;
import java.util.UUID;

/**
 * Entidad JPA que representa la tabla 'users' en PostgreSQL.
 * Se utiliza el nombre 'UserEntity' para evitar conflictos con 'org.springframework.security.core.userdetails.User'
 * o la clase autogenerada 'User' de Protobuf.
 */
@Entity
@Table(name = "users", indexes = {
        // Índice explícito para acelerar las búsquedas por email, ya que es único
        @Index(name = "idx_user_email", columnList = "email", unique = true)
})
public class UserEntity {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID) // Delega la generación de UUID a JPA/Hibernate
    @Column(name = "id", updatable = false, nullable = false)
    private UUID id;

    @NotBlank(message = "El nombre no puede estar vacío")
    @Column(name = "nombre", nullable = false, length = 100)
    private String nombre;

    @NotBlank(message = "El email no puede estar vacío")
    @Email(message = "El formato del email debe ser válido")
    @Column(name = "email", nullable = false, unique = true, length = 150)
    private String email;

    @NotBlank(message = "El hash de la contraseña es requerido")
    @Column(name = "password_hash", nullable = false)
    private String passwordHash;

    @NotNull(message = "El rol es requerido")
    @Enumerated(EnumType.STRING) // Guardamos como String (CLIENTE/ADMIN) para que sea legible en DB, en lugar de 0 o 1
    @Column(name = "rol", nullable = false, length = 20)
    private Role rol;

    @Column(name = "telefono", length = 20)
    private String telefono;

    @CreationTimestamp // Hibernate gestiona automáticamente esta fecha en la inserción
    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    // Constructor vacío requerido por JPA
    public UserEntity() {
    }

    // Getters y Setters
    public UUID getId() {
        return id;
    }

    public void setId(UUID id) {
        this.id = id;
    }

    public String getNombre() {
        return nombre;
    }

    public void setNombre(String nombre) {
        this.nombre = nombre;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public String getPasswordHash() {
        return passwordHash;
    }

    public void setPasswordHash(String passwordHash) {
        this.passwordHash = passwordHash;
    }

    public Role getRol() {
        return rol;
    }

    public void setRol(Role rol) {
        this.rol = rol;
    }

    public String getTelefono() {
        return telefono;
    }

    public void setTelefono(String telefono) {
        this.telefono = telefono;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Instant createdAt) {
        this.createdAt = createdAt;
    }
}
