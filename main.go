package main

import (
	"bytes"
	"encoding/json"
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

	var updateRequest map[string]interface{}
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato JSON inválido"})
		return
	}

	query := `UPDATE medical_records SET description = $1, id = $2`

	args := []interface{}{
		updateRequest["description"],
		patientID,
	}

	queryRequest := map[string]interface{}{
		"sql":  query,
		"args": args,
	}

	queryBody, err := json.Marshal(queryRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al preparar la consulta"})
		return
	}

	queryServiceURL := "http://localhost:8001/query"
	resp, err := http.Post(queryServiceURL, "application/json", bytes.NewBuffer(queryBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al conectar con el servicio de consulta"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Error al actualizar los datos del paciente"})
		return
	}

	var queryResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&queryResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la respuesta del servicio de consulta"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Datos actualizados exitosamente"})
}
