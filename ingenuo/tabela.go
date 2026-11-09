// Arquivo para difinição da estrutura da Tabela
// A Tabela é usada para armazenar os chamados e as seções de forma organizada.
// A estrutura de dados escolhida é a Spatial Hashing
// Esse arquivo foi feito se baseando no código da atividade 3 e na aula 10
package main

import (
	"fmt"
)

// Tabela representa uma tabela de hash espacial para armazenar as posiçõesdos chamados dentro de seções.
// posicoes é um slice de slices de Par, onde cada Par contém uma chave e uma lista de ids de chamados.
type Tabela struct {
	posicoes  [][]Par
}

// Par representa um par de chave e lista de ids de chamados
type Par struct {
	chave string
	chamadosId []int
}

// Retorna uma nova Tabela com m chaves, todas vazias e livres.
func Nova(m int) *Tabela {
	return &Tabela{
		posicoes:   make([][]Par, m),
	}
}

func (t *Tabela) m() int {
	return len(t.posicoes)
}

// hash calcula qual retorna a chave de um chamado na tabela para o chamado com id k.
// A chave é calculada a partir da posição do chamado e do tamanho das seções da tabela.
func (t *Tabela) hash(k Chamado) string {
	x := k.X / TAM_SECAO
	y := k.Y / TAM_SECAO
	h := fmt.Sprintf("(%d,%d)", x, y)
	return h
}

// func (t *Tabela) busca(k Chamado) (int, bool) {
// }

// func (t *Tabela) insere(k Chamado) {
// }

// func (t *Tabela) remove(k Chamado) {
// }
