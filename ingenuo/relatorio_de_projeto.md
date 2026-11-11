# Relatório de projeto: Eduardo F. Pressutto e Lucas W. F. Pepes

## Achar os chamados próximos de uma ambulância
**Estrutura:** Foi feita uma divisão do mapa em seções, atualmente de 25x25, e em cada seção tem um slice com a posição dos chamados, no slice geral de chamados, que estão nela. Na hora de escolher o chamado é percorrido um 3x3 com a seção atual da ambulância no meio. É escolhido o chamado válido com melhor custo-benefício.  
**Custo por turno:** São analisado os chamados presentes no 3x3 de seções para cada ambulância. O que daria um custo de O(C), onde C é o número de chamados nas 9 seções, para cada ambulância. No pior caso, todos os chamados estão presentes e seria O(N) para cada ambulância.  
**Alternativa descartada:** Usar apenas o slice de chamados e percorrê-lo todas as vezes inteiramente, porque teria custo O(N) para cada ambulância sempre e principalmente no cenário de estresse isso consumiria muito tempo com risco de ultrapassar o limite.  

## Descartar os chamados expirados
**Estrutura:**  
**Custo por turno:**  
**Alternativa descartada:** ..., porque ...  

## Remover um chamado recolhido, só com o indentificador
**Estrutura:**  
**Custo por turno:**  
**Alternativa descartada:** ..., porque ...  

## Evitar que duas ambulâncias sigam para o mesmo chamado
**Estrutura:**  
**Custo por turno:**  
**Alternativa descartada:** ..., porque ...  

## Achar o hospital mais próximo de uma ambulância ocupada
**Estrutura:**  
**Custo por turno:**  
**Alternativa descartada:** ..., porque ...  