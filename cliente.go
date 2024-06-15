package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	//coordinatorIP = "<188.188.180.185:8001>" // Cambia esto a la IP del coordinador
	dataURL = "https://raw.githubusercontent.com/educaba123/ta4-archivos/main/student_data.csv"
)

// bitacota de Ips
var addrs []string
var hostaddr string

type Point struct {
	AssessedValue float64
	SaleAmount    float64
}

func leerDatosDesdeCSV(url string) ([]Point, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var points []Point
	for _, record := range records[1:] {
		assessedValue, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			return nil, err
		}
		saleAmount, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, err
		}
		points = append(points, Point{AssessedValue: assessedValue, SaleAmount: saleAmount})
	}
	return points, nil
}

func linearRegression(points []Point) (m, b float64) {
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(points))
	for _, p := range points {
		sumX += p.AssessedValue
		sumY += p.SaleAmount
		sumXY += p.AssessedValue * p.SaleAmount
		sumX2 += p.AssessedValue * p.AssessedValue
	}
	if denom := (n*sumX2 - sumX*sumX); denom != 0 {
		m = (n*sumXY - sumX*sumY) / denom
		b = (sumY - m*sumX) / n
	}
	return
}

func descubrirIP() string {
	//interfaz de red
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces { //Interfaz de red
		if strings.HasPrefix(i.Name, "Ethernet") { //"Ethernet"
			//solo aquellos que son Ethernet
			addrs, _ := i.Addrs()
			for _, addr := range addrs {
				switch t := addr.(type) {
				case *net.IPNet:
					if t.IP.To4() != nil {
						return t.IP.To4().String() //retornamos la IP Ethernet V4
					}
				}
			}
		}
	}
	return "127.0.0.1"
}

func main() {
	//descubrir la IP
	hostaddr = descubrirIP()
	fmt.Println("Mi IP: ", hostaddr)

	br := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese la IP de nodo remoto: ")
	remoteIP, _ := br.ReadString('\n')
	remoteIP = strings.TrimSpace(remoteIP)

	points, err := leerDatosDesdeCSV(dataURL)
	if err != nil {
		fmt.Println("Error al leer los datos desde la URL:", err)
		return
	}

	m, b := linearRegression(points)

	con, err := net.Dial("tcp", remoteIP)
	if err != nil {
		fmt.Println("Error al conectar con el coordinador:", err)
		return
	}
	defer con.Close()

	fmt.Fprintf(con, "%f %f\n", m, b)
	fmt.Println("Resultados enviados al coordinador:", m, b)
}