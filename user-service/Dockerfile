# Etapa 1: Construcción (Build)
FROM maven:3.8.4-openjdk-17-slim AS build
WORKDIR /app

# Copiar el archivo de configuración de Maven y descargar dependencias
COPY pom.xml .
RUN mvn dependency:go-offline -B

# Copiar el código fuente y construir el JAR
COPY src ./src
RUN mvn clean package -DskipTests

# Etapa 2: Ejecución (Runtime)
FROM eclipse-temurin:17-jre
WORKDIR /app

# Copiar el JAR generado desde la etapa de construcción
COPY --from=build /app/target/*.jar app.jar

# Exponer el puerto gRPC configurado
EXPOSE 9090

# Ejecutar la aplicación
ENTRYPOINT ["java", "-jar", "app.jar"]