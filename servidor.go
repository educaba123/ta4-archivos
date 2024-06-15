package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	port = ":8001"
)

type Point struct {
	AssessedValue float64
	SaleAmount    float64
}

var hostaddr string

func handleConnection(con net.Conn, wg *sync.WaitGroup, results chan<- Point) {
	defer wg.Done()
	defer con.Close()
	var m, b float64
	_, err := fmt.Fscanf(con, "%f %f\n", &m, &b)
	if err != nil {
		fmt.Println("Error al leer los datos:", err)
		return
	}
	results <- Point{AssessedValue: m, SaleAmount: b}
}

func descubrirIP() string {
	//interfaz de red
	ifaces, _ := net.Interfaces()
	for _, i := range ifaces { //Interfaz de red
		if strings.HasPrefix(i.Name, "Wi-Fi") { //"Ethernet"
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
	// Pedir al usuario el número de clientes a esperar
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el número de clientes que desea recibir: ")
	numClientsStr, _ := reader.ReadString('\n')
	numClientsStr = strings.TrimSpace(numClientsStr)
	numClients, err := strconv.Atoi(numClientsStr)
	if err != nil {
		fmt.Println("Número inválido, usando valor por defecto de 1")
		numClients = 1
	}

	hostaddr = descubrirIP()
	fmt.Println("Mi IP: ", hostaddr)

	fmt.Println("Nodo Coordinador escuchando....")
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err.Error())
		os.Exit(1)
	}
	defer ln.Close()

	var wg sync.WaitGroup
	results := make(chan Point, numClients)

	for i := 0; i < numClients; i++ {
		con, err := ln.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión:", err.Error())
			continue
		}
		wg.Add(1)
		go handleConnection(con, &wg, results)
	}

	wg.Wait()
	close(results)

	var totalM, totalB float64
	for result := range results {
		totalM += result.AssessedValue
		totalB += result.SaleAmount
	}

	fmt.Printf("Resultados combinados - Pendiente: %.2f, Intercepto: %.2f\n", totalM/float64(numClients), totalB/float64(numClients))
}
