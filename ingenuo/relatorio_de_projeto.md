# Relatório de projeto: Eduardo F. Pressutto e Lucas W. F. Pepes

## Achar os chamados próximos de uma ambulância
**Estrutura:** Foi feita uma divisão do mapa em seções, atualmente de 25x25, e em cada seção tem um slice com a posição dos chamados, no slice geral de chamados, que estão nela. Na hora de escolher o chamado é percorrido um 3x3 com a seção atual da ambulância no meio. É escolhido o chamado válido com melhor custo-benefício.  
**Custo por turno:** São analisado os chamados presentes no 3x3 de seções para cada ambulância. O que daria um custo de O(c), onde c é o número de chamados nas 9 seções, para cada ambulância. No pior caso, todos os chamados estão presentes e seria O(n) para cada ambulância.  
**Alternativa descartada:** Usar apenas o slice de chamados e percorrê-lo todas as vezes inteiramente, porque teria custo O(n) para cada ambulância sempre e principalmente no cenário de estresse isso consumiria muito tempo com risco de ultrapassar o limite.  

## Descartar os chamados expirados
**Estrutura:** Um slice que vai do turno 0 até o último e contém a posição do chamados, no slice de chamados, que expiram naquele turno. Foi adicionado também um atributo no Chamado que guarda se expirou ou não. Quando chega um chamado ele é inserido na posição correspondete ao seu turno de expiração. Na hora do descarte eles são descartados da hash e expirou = true. Na hora da busca os marcados como expirados são removidos das suas seções.  
**Custo por turno:** O(c), onde c é a quantidade de chamados que expiram no turno.  
**Alternativa descartada:** Usar o slice de chamados e comparar o prazo de cada um com o turno atual, porque o custo seria O(n).  

## Remover um chamado recolhido, só com o indentificador
**Estrutura:** Tabela Hash com os indentificadores representando a posição dos chamados no slice de chamados. Quando um chamado chega ele é inserido passando seu ID como chave e a posição como valor. Quando chega o ID do chamado recolhido é feito uma busca na tabela hash e se obtém a posição dele, o chamado é removido da tabela e seu atributo expirou recebe true.  
**Custo por turno:**  No caso médio é O(1) e no pior caso é O(n) por chamado recolhido, o custo das operações de busca e de remoção se somam, mas não alteram o O().  
**Alternativa descartada:** Usar o slice de chamados e comparar o ID de cada um com o ID recebido, porque o custo seria O(n) por chamado recolhido.  

## Evitar que duas ambulâncias sigam para o mesmo chamado
**Estrutura:** Atributo Atribuido adicionado no Chamado. Quando um chamado vira alvo de alguma ambulância, Atribuido recebe true, quando ele para de ser alvo vira false.  
**Custo por turno:**  O custo é O(1) por ambulância, na hora da ambulância verificar se o alvo é válido apenas olha o valor de Atribuido, se for true não é válido e duas não seguem para o mesmo chamado.  
**Alternativa descartada:** Verificar os valores no slice de alvos, não podendo ser igual a nenhum, porque o custo seria O(a) por ambulância, onde a é a quantidade de ambulâncias. O slice de alvos ainda foi utilizado, mas com outro objetivo.  

## Achar o hospital mais próximo de uma ambulância ocupada
**Estrutura:** Slice de hospitais, percorre todos os hospitais comparando a distância da ambulância até eles. Como o máximo de hospitais é baixo (20 hospitais no cenário de estresse), percorrer o slice inteiro não custa tanto.  
**Custo por turno:** O custo é O(h) por ambulância, onde h é a quantidade de hospitais. Sendo no máximo 20 nos cenários disponíveis.  
**Alternativa descartada:** Adicionar um slice (tamanho igual a quantidade de ambulâncias) que quarda o número do hospital mais próximo de cada ambulância, esses valores seriam associados no começo da partida, porque o número de hospitais era baixo, então achamos que não valeria a pena um slice específico só para isso, mas talvez alteraremos isso na parte 2.  