package main

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func WeatherReport(ciudad string) string {
	var result string
	// URL del archivo que deseas descargar
	url := "https://ssl.smn.gob.ar/dpd/zipopendata.php?dato=tiepre"

	// Realizar una solicitud HTTP GET para obtener el archivo
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error al realizar la solicitud HTTP:", err)
		return ""
	}
	defer response.Body.Close()

	// Verificar el código de respuesta HTTP
	if response.StatusCode != http.StatusOK {
		fmt.Println("Respuesta del servidor no válida:", response.Status)
		return ""
	}

	// Obtener el nombre del archivo del encabezado "Content-Disposition"
	zipFileName := getFilenameFromHeader(response.Header)

	// Si el encabezado "Content-Disposition" no contiene el nombre del archivo, utilizar el último segmento de la URL como nombre
	if zipFileName == "" {
		zipFileName = getFilenameFromURL(url)
	}

	// Crear el archivo en el sistema de archivos
	outFile, err := os.Create(zipFileName)
	if err != nil {
		fmt.Println("Error al crear el archivo:", err)
		return ""
	}
	defer outFile.Close()

	// Copiar el contenido de la respuesta HTTP al archivo local
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		fmt.Println("Error al copiar el contenido:", err)
		return ""
	}

	//fmt.Printf("Archivo descargado como: %s\n", zipFileName)

	// Abrir el archivo ZIP
	zipFile, err := zip.OpenReader(zipFileName)
	if err != nil {
		fmt.Println("Error al abrir el archivo ZIP:", err)
		return ""
	}
	defer zipFile.Close()

	// Encontrar el primer archivo dentro del ZIP
	var firstFile *zip.File
	for _, file := range zipFile.File {
		firstFile = file
		break
	}

	if firstFile == nil {
		fmt.Println("No se encontraron archivos dentro del ZIP.")
		return ""
	}

	// Crear el archivo en el sistema de archivos
	outFile, err = os.Create(firstFile.Name)
	if err != nil {
		fmt.Println("Error al crear el archivo:", err)
		return ""
	}
	defer outFile.Close()

	// Abrir el archivo dentro del ZIP
	rc, err := firstFile.Open()
	if err != nil {
		fmt.Println("Error al abrir el archivo dentro del ZIP:", err)
		return ""
	}
	defer rc.Close()

	// Copiar el contenido del archivo dentro del ZIP al archivo en el sistema de archivos
	_, err = io.Copy(outFile, rc)
	if err != nil {
		fmt.Println("Error al copiar el contenido del archivo:", err)
		return ""
	}

	//fmt.Printf("Archivo %s extraído exitosamente.\n", firstFile.Name)

	// Nombre del archivo CSV que deseas abrir
	csvFileName := firstFile.Name

	// Abrir el archivo CSV
	file, err := os.Open(csvFileName)
	if err != nil {
		fmt.Println("Error al abrir el archivo CSV:", err)
		return ""
	}
	defer file.Close()

	// Crear un lector CSV con punto y coma como delimitador
	reader := csv.NewReader(file)
	reader.Comma = ';'

	var i int16
	var found bool

	found = false
	// Leer y procesar las filas del archivo CSV
recordsloop:
	for {
		// Leer una fila del archivo CSV
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				// Fin del archivo
				break
			}
			fmt.Println("Error al leer una fila del archivo CSV:", err)
			return ""
		}

		// Procesar los campos de la fila (en este ejemplo, simplemente los imprimimos)
		result = ""
		i = 0
		for _, field := range record {
			if i == 0 {
				if strings.Compare(strings.ToLower(strings.Trim(ciudad, " \r\n")), strings.ToLower(strings.Trim(field, " \r\n"))) == 0 {
					//fmt.Printf("%s\t", field)
					result = strings.Trim(field, " \r\n") + "\n"
					found = true
				}
			} else {
				if found {
					//fmt.Printf("%s\t", field)

					//TODO El indice i indica la posicion de la columa
					//dentro del registro, completar el código para
					//acceder a las columnas 2 (hora), 3 (cielo),
					//4 (velocidad del viento), 8 (direccion del viento)
					//y 9 (presion atmosferica)
					switch i {
					case 1:
						result = result + "Fecha: " + field + "\n"

						//case 2:
						//	result = result + "Hora: " + field + "\n"

					}
				}
			}
			i++
		}

		if found {
			break recordsloop
		}
	}
	return result

}

// Función para obtener el nombre de archivo del encabezado "Content-Disposition"
func getFilenameFromHeader(header http.Header) string {
	disposition := header.Get("Content-Disposition")
	if disposition != "" {
		_, params, err := mime.ParseMediaType(disposition)
		if err == nil {
			return params["filename"]
		}
	}
	return ""
}

// Función para obtener el nombre de archivo del último segmento de la URL
func getFilenameFromURL(url string) string {
	_, file := filepath.Split(url)
	return file
}
