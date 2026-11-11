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
	"bot/tabela" // Pacote com a tabela
	"bufio"
	"fmt"
	"math" // Para fazer a potência
	"os"
	"strconv"
	"strings"
)

type Chamado struct {
	ID        int
	X, Y      int
	Valor     int
	Prazo     int
	Expirou   bool // Marca se expirou ou foi recolhido
	Atribuido bool // Marca se uma ambulância já está indo para ele
	DistHosp  int  // Distância até o hospital mais próximo
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

	totalTurnos     int // Total de turnos da partida
	largura, altura int // Dimensões do mapa

	indice *tabela.Tabela // Indice da tabela hash

	// Slice para armazenar os chamados no turno em que eles expiram
	// Cada posição é um turno com um slice dos chamados que expiram nesse turno
	expiraEm [][]int

	// Slice para armazenar o chamado alvo de cada ambulância
	alvos []int

	// Seções para dividir o mapa em blocos TAM_SECAO x TAM_SECAO
	secoes          [][][]int // Seções do mapa, X, Y, posicão dos chamados no slice de chamados
	linhas, colunas int       // Quantidade de linhas e colunas na divisão do mapa
)

// Tamanho das seções
// 25 foi escolhido após alguns testes
// Talvez fique melhor se mudar de acordo com o tamanho do mapa
const TAM_SECAO = 25

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
	totalTurnos = inteiro(c[3])

	// Calcula a quantidade de linhas e colunas de seções do mapa
	linhas = altura / TAM_SECAO
	colunas = largura / TAM_SECAO

	h := inteiro(lerCampos()[1])
	for i := 0; i < h; i++ {
		c := lerCampos()
		hospitais = append(hospitais, Hospital{inteiro(c[0]), inteiro(c[1])})
	}
	k := inteiro(lerCampos()[1])
	alvos = make([]int, k) // Inicializa o slice dos alvos das ambulâncias
	for i := 0; i < k; i++ {
		lerCampos()   // id x y
		alvos[i] = -1 // Inicialmente todas sem alvo (-1)
	}
	b := inteiro(lerCampos()[1])
	for i := 0; i < b; i++ {
		lerCampos() // x1 y1 x2 y2
	}
	lerCampos() // FIM
}

// Função não utilizada
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

// Calcula em qual seção do mapa o ponto (x, y) está
func calcularSecao(y, x int) (int, int) {
	linha := y / TAM_SECAO
	coluna := x / TAM_SECAO

	// Garantia que a coordenada fique em uma seção no mapa, protege contra as cordenadas nas pontas do mapa
	if linha >= linhas {
		linha = linhas - 1
	}
	if coluna >= colunas {
		coluna = colunas - 1
	}

	return linha, coluna
}

// removerExpirados mantém só os chamados com prazo maior ou igual ao turno.
// Acessa o slice expiraEm[turno-1] para obter todos que expiram nesse turno, marca Expirou como true e remove da tabela hash.
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

// acharHospProx devolve o índice do hospital mais próximo do ponto (x, y) e a distância até ele.
// Percorre todos os hospitais, pois o número máximo de hospitais é 20, então não fica tão caro.
func acharHospProx(x, y int) (int, int) {
	melhor := 0
	distanciaMelhor := distancia(x, y, hospitais[melhor].X, hospitais[melhor].Y)
	for i, h := range hospitais {
		distancia := distancia(x, y, h.X, h.Y)
		if distancia < distanciaMelhor {
			melhor = i
			distanciaMelhor = distancia
		}
	}
	return melhor, distanciaMelhor
}

// custoBenefício calcula o quanto vale a pena ir até um chamado
// Considera o valor dele, a distância até o chamado e a distância até o hospital mais próximo
// Fomos testando até achar o expoente 0.7 como um bom valor e o +1 evita divisão por 0
func custoBeneficio(pos, dist int) float64 {
	return math.Pow(float64(chamados[pos].Valor), 0.7) / float64(dist+chamados[pos].DistHosp+1)
}

// daPraChegar verifica se é possível ir até o chamado antes dele expirar
func daPraChegar(x, y, pos, turno int) (bool, int) {
	dist := distancia(x, y, chamados[pos].X, chamados[pos].Y)
	da := dist <= chamados[pos].Prazo-turno+1
	return da, dist
}

// acharMelhorChamado retorna o melhor chamado para a ambulância
func acharMelhorChamado(x, y, turno int) int {
	// Calcula a seção do mapa em que a ambulância está
	linhaAmbulancia, colunaAmbulancia := calcularSecao(y, x)
	melhor, melhorCB := -1, -1.0 // Melhor chamado e seu custo-benefício

	// Faz uma varredura nas seções ao redor da ambulância, incluindo a atual
	for offy := -1; offy <= 1; offy++ {
		linhaAtual := linhaAmbulancia + offy
		// Verifica se a linha existe
		if linhaAtual < 0 || linhaAtual >= linhas {
			continue
		}
		for offx := -1; offx <= 1; offx++ {
			colunaAtual := colunaAmbulancia + offx
			// Verifica se a coluna existe
			if colunaAtual < 0 || colunaAtual >= colunas {
				continue
			}
			// Varre a seção atual e atualiza o melhor chamado e seu custo-benefício
			melhor, melhorCB = varrerSecao(x, y, turno, linhaAtual, colunaAtual, melhor, melhorCB)
		}
	}
	return melhor
}

// Varre uma seção do mapa e returna o melhor chamado e seu custo-benefício para a posição (x, y)
func varrerSecao(x, y, turno, linha, coluna, melhor int, melhorCB float64) (int, float64) {
	secaoAtual := secoes[linha][coluna] // Pega o slice de chamados da seção atual
	for i := 0; i < len(secaoAtual); i++ {
		pos := secaoAtual[i]
		chamado := chamados[pos]
		// Verifica se o chamado expirou, se sim, remove do slice da seção atual trocando com o último
		if chamado.Expirou {
			secaoAtual[i] = secaoAtual[len(secaoAtual)-1] // Troca
			secaoAtual = secaoAtual[:len(secaoAtual)-1]   // Remove
			i--
			continue
		} else {
			// Verifica se dá para chegar no chamado antes de expirar e se ele já não está atribuído a outra ambulância
			da, dist := daPraChegar(x, y, pos, turno)
			if !chamado.Atribuido && da {
				// Calcula o custo-benefício do chamado e caso ele seja melhor, atualiza
				cb := custoBeneficio(pos, dist)
				if cb > melhorCB {
					melhor = pos
					melhorCB = cb
				}
			}
		}
	}
	// Atualiza o slice da seção atual
	secoes[linha][coluna] = secaoAtual
	return melhor, melhorCB
}

func jogarTurno(turno int, ambulancias []Ambulancia) {
	for _, a := range ambulancias {
		mov := "P"
		posAlvo := alvos[a.ID] // Pega o alvo atual da ambulância
		if a.Livre {
			// Se tiver alvo, verifica se ele ainda é válido
			if posAlvo != -1 {
				dist := distancia(a.X, a.Y, chamados[posAlvo].X, chamados[posAlvo].Y)
				if chamados[posAlvo].Expirou || dist > chamados[posAlvo].Prazo-turno+1 {
					chamados[posAlvo].Atribuido = false
					posAlvo = -1
				}
			}

			// Se não tiver alvo, procura o melhor para ela
			if posAlvo == -1 {
				posAlvo = acharMelhorChamado(a.X, a.Y, turno)
				// Se achar, marca o chamado como atribuído
				if posAlvo != -1 {
					chamados[posAlvo].Atribuido = true
				}
			}

			// Atualiza o alvo da ambulância
			alvos[a.ID] = posAlvo

			// Se tiver alvo, anda na direção dele, senão vai para o hospital mais próximo
			if posAlvo != -1 {
				mov = direcao(a.X, a.Y, chamados[posAlvo].X, chamados[posAlvo].Y)
			} else {
				h, _ := acharHospProx(a.X, a.Y)
				mov = direcao(a.X, a.Y, hospitais[h].X, hospitais[h].Y)
			}
		} else {
			// Libera o alvo da ambulância para caso ela tenha recolhido outro chamado no caminho
			if posAlvo != -1 {
				chamados[posAlvo].Atribuido = false
			}
			// Volta ao estado sem alvo e vai para o hospital mais próximo
			alvos[a.ID] = -1
			melhor, _ := acharHospProx(a.X, a.Y)
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

	// Preprocessamento
	lerInicio()

	// Cria a tabela hash para armazenar as posições dos chamados
	indice = tabela.Nova(1000003)

	// Inicializa as seções do mapa
	secoes = make([][][]int, linhas)
	for i := 0; i < linhas; i++ {
		secoes[i] = make([][]int, colunas)
	}

	// Inicializa o slice de chamados que expiram em cada turno
	expiraEm = make([][]int, totalTurnos+1) // +1 para incluir o último turno

	fmt.Fprintln(escritor, "PRONTO")
	escritor.Flush()
	// Término do preprocessamento

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
				false, -1})
			pos := len(chamados) - 1 // Pega a posição do chamado no slice
			chamado := chamados[pos] // Pega o chamado recém-adicionado

			// Insere no slice para expiração de acordo com o prazo + 1, pois o chamado expira no turno seguinte ao prazo
			// Só faz isso se o prazo for menor que o total de turnos, para não estourar o slice
			if chamado.Prazo < totalTurnos {
				expiraEm[chamado.Prazo+1] = append(expiraEm[chamado.Prazo+1], pos)
			}

			// Insere na tabela hash
			indice.Insere(chamado.ID, pos)

			// Insere na sua respectiva seção do mapa
			linha, coluna := calcularSecao(chamado.Y, chamado.X)
			secoes[linha][coluna] = append(secoes[linha][coluna], pos)

			// Inicializa a distância até o hospital mais próximo
			_, chamados[pos].DistHosp = acharHospProx(chamado.X, chamado.Y)
		}

		r := inteiro(lerCampos()[1]) // RECOLHIDOS r
		for i := 0; i < r; i++ {
			id := inteiro(lerCampos()[0])
			// Busca a posição do chamado no slice a partir do id na tabela hash
			pos, ok := indice.Busca(id)
			// Remove o chamado da tabela hash e marca como expirado
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
