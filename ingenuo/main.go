// Bot ingênuo: ponto de partida do trabalho prático.
//
// Este bot implementa o protocolo inteiro e joga de forma simples:
//
//   - guarda os chamados abertos num slice;
//   - a cada turno, cada ambulância livre procura, por varredura linear,
//     o chamado mais próximo e anda na direção dele;
//   - cada ambulância ocupada anda na direção do hospital mais próximo.
//
// Ele não verifica se o chamado ainda pode ser alcançado dentro do prazo,
// não impede que duas ambulâncias sigam para o mesmo chamado e ignora os
// bloqueios da etapa 2. No cenário de estresse, a varredura linear estoura
// o limite de tempo por turno.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"bot/tabela"
)

type Chamado struct {
	ID    int
	X, Y  int
	Valor int
	Prazo int
	Expirou bool
	Atribuido bool
	HospProx int
	DistHosp int
}

type Ambulancia struct {
	ID    int
	X, Y  int
	Livre bool
}

type Hospital struct {
	X, Y int
}

var (
	leitor   *bufio.Scanner
	escritor *bufio.Writer
	
	hospitais []Hospital
	chamados  []Chamado
	
	totalTurnos int
	largura, altura int

	indice *tabela.Tabela // Indice da tabela hash

	// Slice para armazenar os chamados no turno em que eles expiram
	expiraEm [][]int 

	//
	alvos[]int

	// Seções para dividir o mapa em blocos TAM_SECAO x TAM_SECAO
	secoes [][][]int // Seções do mapa, X, Y, IDs dos chamados
	linhas, colunas int // Quantidade de linhas e colunas na divisão do mapa 
)

// Tamanho das seções
const TAM_SECAO = 50

// lerCampos lê a próxima linha da entrada e a divide em campos.
// Devolve nil quando a entrada termina.
func lerCampos() []string {
	if !leitor.Scan() {
		return nil
	}
	return strings.Fields(leitor.Text())
}

func inteiro(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic("número inválido na entrada: " + s)
	}
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func distancia(x1, y1, x2, y2 int) int {
	return abs(x1-x2) + abs(y1-y2)
}

// direcao devolve o movimento que aproxima (x, y) de (alvoX, alvoY),
// andando primeiro na horizontal.
func direcao(x, y, alvoX, alvoY int) string {
	switch {
	case x < alvoX:
		return "L"
	case x > alvoX:
		return "O"
	case y < alvoY:
		return "S"
	case y > alvoY:
		return "N"
	}
	return "P"
}

// lerInicio lê a mensagem de início. Este bot só usa os hospitais; os
// bloqueios são lidos e descartados.
func lerInicio() {
	c := lerCampos() // INICIO largura altura turnos limite_turno limite_preparo equipe num_equipes
	
	largura = inteiro(c[1])
	altura = inteiro(c[2])
	linhas = altura / TAM_SECAO
	colunas = largura / TAM_SECAO

	totalTurnos = inteiro(c[3])
	h := inteiro(lerCampos()[1])
	for i := 0; i < h; i++ {
		c := lerCampos()
		hospitais = append(hospitais, Hospital{inteiro(c[0]), inteiro(c[1])})
	}
	k := inteiro(lerCampos()[1])
	alvos = make([]int, k)
	for i := 0; i < k; i++ {
		lerCampos() // id x y
		alvos[i] = -1
	}
	b := inteiro(lerCampos()[1])
	for i := 0; i < b; i++ {
		lerCampos() // x1 y1 x2 y2
	}
	lerCampos() // FIM
}

// removerPorID procura o chamado no slice e o remove, trocando-o pelo
// último elemento. Custo: proporcional ao número de chamados abertos.
// func removerPorID(id int) {
// 	for i := range chamados {
// 		if chamados[i].ID == id {
// 			chamados[i] = chamados[len(chamados)-1]
// 			chamados = chamados[:len(chamados)-1]
// 			return
// 		}
// 	}
// }

// 
func calcularSecao(y, x int) (int, int) {
	linha := y/TAM_SECAO
	coluna := x/TAM_SECAO

	return linha, coluna
}

// removerExpirados mantém só os chamados com prazo maior ou igual ao turno.
func removerExpirados(turno int) {
	chamadosExpirados := expiraEm[turno-1]
	if turno-1 >= 1 {
		for _, pos := range chamadosExpirados {
			indice.Remove(chamados[pos].ID)
			chamados[pos].Expirou = true
		}
		expiraEm[turno-1] = nil
	} 
}

//
func acharHospProx(x, y int) (int, int) {
	melhor := 0
	distanciaMelhor := distancia(x, y, hospitais[melhor].X, hospitais[melhor].Y)
	for i, h := range hospitais{
		distancia := distancia(x, y, h.X, h.Y)
		if distancia < distanciaMelhor {
			melhor = i
			distanciaMelhor = distancia
		}
	}
	return melhor, distanciaMelhor
}

func custoBeneficio(pos, dist int) float64 {
	return float64(chamados[pos].Valor)/float64(dist + chamados[pos].DistHosp + 1)
}

func daPraChegar(x, y, pos, turno int) (bool, int){
	dist := distancia(x, y, chamados[pos].X, chamados[pos].Y)
	da := dist <= chamados[pos].Prazo - turno + 1
	return da, dist
}

func acharMelhorChamado(x, y, turno int) int {
	linhaAmbulancia, colunaAmbulancia := calcularSecao(y, x)
	melhor, melhorCB := -1, -1.0

	for offy := -1; offy <= 1; offy++ {
		linhaAtual := linhaAmbulancia + offy
		if (linhaAtual < 0 || linhaAtual > linhas) {
			continue
		}
		for offx := -1; offx <=1; offx++ {
			colunaAtual := colunaAmbulancia + offx
			if (colunaAtual < 0 || colunaAtual > colunas) {
				continue
			}
			melhor, melhorCB = varrerSecao(x, y, turno, linhaAtual, colunaAtual, melhor, melhorCB)
		}
	}
	return melhor
}

func varrerSecao(x, y, turno, linha, coluna, melhor int, melhorCB float64) (int, float64) {
	secaoAtual := secoes[linha][coluna]
	for i := 0; i < len(secaoAtual); i++ {
		pos := secaoAtual[i]
		chamado := chamados[pos]
		if chamado.Expirou {
			secaoAtual[i] = secaoAtual[len(secaoAtual)-1]
			secaoAtual = secaoAtual[:len(secaoAtual)-1]
			i--
			continue
		} else {
			da, dist := daPraChegar(x, y, pos, turno)
			if !chamado.Atribuido && da {
				cb := custoBeneficio(pos, dist)
				if cb > melhorCB {
					melhor = pos
					melhorCB = cb
				}
			}
		}
	}
	secoes[linha][coluna] = secaoAtual
	return melhor, melhorCB
}

func jogarTurno(turno int, ambulancias []Ambulancia) {
	for _, a := range ambulancias {
		mov := "P"
		posAlvo := alvos[a.ID]
		if a.Livre {
			if posAlvo != -1 {
				d := distancia(a.X, a.Y, chamados[posAlvo].X, chamados[posAlvo].Y)
				if chamados[posAlvo].Expirou || d > chamados[posAlvo].Prazo - turno + 1 {
					chamados[posAlvo].Atribuido = false
					posAlvo = -1
				}
			}

			if posAlvo == -1 {
				posAlvo = acharMelhorChamado(a.X, a.Y, turno)
				if posAlvo != -1 {
					chamados[posAlvo].Atribuido = true
				}
			}

			alvos[a.ID] = posAlvo

			if posAlvo != -1 {
				mov = direcao(a.X, a.Y, chamados[posAlvo].X, chamados[posAlvo].Y)
			} else {
				h, _ := acharHospProx(a.X, a.Y)
				mov = direcao(a.X, a.Y, hospitais[h].X, hospitais[h].Y)
			}
		} else {
			alvos[a.ID] = -1
			melhor := chamados[posAlvo].HospProx
			mov = direcao(a.X, a.Y, hospitais[melhor].X, hospitais[melhor].Y)
		}
		fmt.Fprintln(escritor, a.ID, mov)
	}
	fmt.Fprintln(escritor, "FIM")
	escritor.Flush()
}

func main() {
	leitor = bufio.NewScanner(os.Stdin)
	leitor.Buffer(make([]byte, 1<<16), 1<<20)
	escritor = bufio.NewWriter(os.Stdout)

	lerInicio()

	// Preprocessamento
	// Cria a tabela hash para armazenar as posições dos chamados
	indice = tabela.Nova(2000003)

	// Inicializa as seções do mapa
	secoes = make([][][]int, linhas)
	for i := 0; i < linhas; i++ {
		secoes[i] = make([][]int, colunas)
	}

	// Inicializa o slice de chamados que expiram em cada turno
	expiraEm = make([][]int, totalTurnos+1) // +1 para incluir o último turno

	fmt.Fprintln(escritor, "PRONTO")
	escritor.Flush()

	for {
		c := lerCampos() // TURNO t
		if c == nil {
			return // fim da partida
		}
		turno := inteiro(c[1])

		n := inteiro(lerCampos()[1]) // CHAMADOS n
		for i := 0; i < n; i++ {
			c := lerCampos()
			chamados = append(chamados, Chamado{inteiro(c[0]), inteiro(c[1]), inteiro(c[2]),
												inteiro(c[3]), inteiro(c[4]), false,
												false, -1, -1})
			pos := len(chamados)-1									
			chamado := chamados[pos]

			// Insere no slice para expiração
			if chamado.Prazo < totalTurnos {
				expiraEm[chamado.Prazo+1] = append(expiraEm[chamado.Prazo+1], pos)
			}
			
			// Insere na tabela hash
			indice.Insere(chamado.ID, pos)

			// Insere na sua respectiva seção do mapa
			linha, coluna := calcularSecao(chamado.Y, chamado.X)
			secoes[linha][coluna] = append(secoes[linha][coluna], pos)

			// Inicializa os atributos de hospital mais próximo e distância até esse hospital
			chamados[pos].HospProx, chamados[pos].DistHosp = acharHospProx(chamado.X, chamado.Y)
		}

		r := inteiro(lerCampos()[1]) // RECOLHIDOS r
		for i := 0; i < r; i++ {
			id := inteiro(lerCampos()[0])
			pos, ok := indice.Busca(id)
			if ok {
				chamados[pos].Expirou = true
				indice.Remove(id)
			}
		}
		removerExpirados(turno)

		k := inteiro(lerCampos()[1]) // AMBULANCIAS k
		ambulancias := make([]Ambulancia, k)
		for i := 0; i < k; i++ {
			c := lerCampos()
			ambulancias[i] = Ambulancia{inteiro(c[0]), inteiro(c[1]), inteiro(c[2]), c[3] == "LIVRE"}
		}

		e := inteiro(lerCampos()[1]) // PLACAR e
		for i := 0; i < e; i++ {
			lerCampos()
		}
		lerCampos() // FIM

		jogarTurno(turno, ambulancias)
	}
}
