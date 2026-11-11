// Arquivo para difinição da estrutura da Tabela Hash
// A Tabela é usada para armazenar a posição dos chamados no slice.
// A chaves são os IDs dos chamados hashados
// Esse arquivo foi feito se baseando no código da atividade 3
// As funções são iguais as que estão na tabela.go da atividade 3
package tabela

import (
	"fmt"
	"strings"
)

// Tabela associa chaves (ids dos chamados) a valores (índices no slice de chamados).
// Ocupados indica se a posição da tabela está ocupada ou não.
type Tabela struct {
	chaves  []int
	valores []int
	ocupada []bool
	n       int // número de pares guardados
}

// Nova cria uma nova tabela com m posições.
func Nova(m int) *Tabela {
	return &Tabela{
		chaves:  make([]int, m),
		valores: make([]int, m),
		ocupada: make([]bool, m),
	}
}

// Len devolve o número de pares na tabela.
func (t *Tabela) Len() int {
	return t.n
}

// m devolve o número de posições na tabela.
func (t *Tabela) m() int {
	return len(t.chaves)
}

// misturar = false parece desempenhar melhor ou não mudar nada
const misturar = false

// Faz o hash da chave (ID do chamado)
// Se misturar = true, mistura os bits
func (t *Tabela) hash(k int) int {
	if misturar {
		k = mistura(k)
	}
	h := k % t.m()
	if h < 0 {
		h += t.m()
	}
	return h
}

// mistura embaralha os bits da chave com deslocamentos, ou-exclusivos e
// multiplicações por constantes ímpares grandes, a mesma ideia da FNV-1a.
// Chaves vizinhas, como 1000 e 1001, saem com valores sem relação entre si.
// Os passos e as constantes são os do gerador SplitMix64.
func mistura(k int) int {
	x := uint64(k)
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return int(x >> 1) // descarta o bit de sinal, para o resultado ser positivo
}

// distancia devolve quantos passos a sondagem linear dá da posição de até a
// posição ate, dando a volta no fim do vetor quando ate vem antes de de.
// Com m = 11: distancia(9, 1) = 3, porque a sondagem passa por 9, 10, 0 e 1.
func (t *Tabela) distancia(de, ate int) int {
	return (ate - de + t.m()) % t.m()
}

// Busca devolve o valor associado à chave e true, ou 0 e false se a chave
// não está na tabela. A busca começa em h(k) e anda uma posição por vez até
// achar a chave ou uma posição vazia. Ela sempre termina, porque a Insere
// mantém pelo menos 30% das posições vazias.
func (t *Tabela) Busca(k int) (int, bool) {
	p := t.hash(k)
	for t.ocupada[p] {
		if t.chaves[p] == k {
			return t.valores[p], true
		}
		p = (p + 1) % t.m()
	}
	return 0, false
}

// String mostra as chaves em cada posição, com um ponto nas posições vazias,
// no mesmo formato das tabelas da aula de tabelas de dispersão. Útil para
// depurar com fmt.Println(t).
func (t *Tabela) String() string {
	var b strings.Builder
	for p := 0; p < t.m(); p++ {
		fmt.Fprintf(&b, "%4d", p)
	}
	b.WriteString("\n")
	for p := 0; p < t.m(); p++ {
		if t.ocupada[p] {
			fmt.Fprintf(&b, "%4d", t.chaves[p])
		} else {
			b.WriteString("   .")
		}
	}
	return b.String()
}

// proximoPrimo devolve o menor primo maior ou igual a k.
func proximoPrimo(k int) int {
	for {
		primo := k >= 2
		for d := 2; d*d <= k; d++ {
			if k%d == 0 {
				primo = false
				break
			}
		}
		if primo {
			return k
		}
		k++
	}
}
