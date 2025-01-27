package main

import (
    "bytes"
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // Endpoint para editar los datos del paciente
    r.PUT("/edit/:id", updatePatientData)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8082"
    }

    log.Printf("Servidor corriendo en :%s", port)
    r.Run(":" + port)
}

func updatePatientData(c *gin.Context) {
    patientID := c.Param("id")
    log.Printf("Actualizando registro para el paciente ID: %s", patientID)

    // Leer y loggear el body raw
    bodyBytes, err := ioutil.ReadAll(c.Request.Body)
    if err != nil {
        log.Printf("Error al leer el body: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Error al leer el request"})
        return
    }

    // Restaurar el body
    c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
    
    // Loggear el body raw
    log.Printf("Body raw recibido: %s", string(bodyBytes))

    var updateRequest map[string]interface{}
    if err := c.ShouldBindJSON(&updateRequest); err != nil {
        log.Printf("Error al analizar el JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Formato JSON inválido"})
        return
    }

    // Verificar que description existe en el request
    if _, exists := updateRequest["description"]; !exists {
        log.Printf("Falta el campo requerido: description")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Falta el campo: description"})
        return
    }

    // Corregir la consulta SQL - usar WHERE para el ID
    query := `UPDATE medical_records SET description = $1 WHERE id = $2`

    args := []interface{}{
        updateRequest["description"],
        patientID,
    }

    queryRequest := map[string]interface{}{
        "sql":  query,
        "args": args,
    }

    // Log de la consulta que se va a ejecutar
    queryJSON, _ := json.Marshal(queryRequest)
    log.Printf("Consulta a ejecutar: %s", string(queryJSON))

    queryBody, err := json.Marshal(queryRequest)
    if err != nil {
        log.Printf("Error al preparar la consulta: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
        return
    }

    queryServiceURL := "http://localhost:8001/query"
    resp, err := http.Post(queryServiceURL, "application/json", bytes.NewBuffer(queryBody))
    if err != nil {
        log.Printf("Error al conectar con el servicio de consulta: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al conectar con el servicio de consulta"})
        return
    }
    defer resp.Body.Close()

    // Leer y loggear el body de la respuesta
    respBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error al leer la respuesta: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al leer la respuesta"})
        return
    }

    // Loggear la respuesta
    log.Printf("Respuesta del servicio de consulta: %s", string(respBody))

    // Restaurar el body para poder decodificarlo
    resp.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))

    if resp.StatusCode != http.StatusOK {
        log.Printf("Error al actualizar los datos del paciente, respuesta del servidor: %s", resp.Status)
        c.JSON(resp.StatusCode, gin.H{"error": "Error al actualizar los datos del paciente"})
        return
    }

    var queryResponse map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&queryResponse); err != nil {
        log.Printf("Error al procesar la respuesta del servicio de consulta: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la respuesta del servicio de consulta"})
        return
    }

    log.Printf("Actualización exitosa para el paciente ID: %s", patientID)
    c.JSON(http.StatusOK, gin.H{"message": "Datos actualizados exitosamente"})
}