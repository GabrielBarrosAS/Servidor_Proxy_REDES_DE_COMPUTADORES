/*
==========================================INFORMAÇÕES PARA AVALIAÇÃO==========================================
XX  Passo 1: Encaminhar solicitações e respostas HTTP sem armazenamento em cache                            XX
XX                      Este item foi completamente implementado no trabalho, é                             XX
XX                      possível acessar as páginas web por meio de nosso servidor proxy                    XX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XX  Passo 2: Habilite o servidor TCP para lidar com conexões simultâneas                                    XX
XX                      Este item foi completamente implementado, por meio das técnicas                     XX
XX                      mostradas em sala com a utilização das goroutines. Um exemplo                       XX
XX                      de múltiplos acessos pode ser visto quando acessamos a página                       XX
XX                      do sigaa enquanto estamos navegando na página de Kurose                             XX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XX  Passo 3: Habilitar cache                                                                                XX
XX                      Este item foi completamente implementado, nossos registros de                       XX
XX                      requisições ficam salvos na lista global cache em tempo de                          XX
XX                      execução, e persitem no arquivo cache.txt presente dentro da                        XX
XX                      pasta app, fazendo a manipulação de arquivos em go                                  XX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XX  Passo 4: Fazer com que os itens do cache expirem                                                        XX
XX                      Este item foi completamente implementado, na apresentação é                         XX
XX                      possível identificar a abordagem e tratamento de tempo que                          XX
XX                      foram abordados na realização do trabalho prático 1                                 XX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
XX  Passo 5: Modifique a resposta HTTP                                                                      XX
XX                      Não conseguimos implementar este item em sua totalidade até                         XX
XX                      o prazo final de entrega, assim optamos por remover em                              XX
XX                      totalidade a programação relacionada ao passo 5                                     XX
XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
*/
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "8888"
	CONN_TYPE = "tcp"
)

type no struct {
	url       string
	nome      string
	diretorio string
	tempo     int //Estrutura elaborada para registrar cada requisição na cache
	content   []byte
	ip        string
}

var cache []no

func main() {
	carregarCache() //Carregando os registrps que foram criados em execuções anteriores
	servidor()
}

func servidor() {
	var addr, _ = net.ResolveTCPAddr("tcp", ":"+CONN_PORT)
	server, err := net.ListenTCP(CONN_TYPE, addr) // Servidor escutando no localhost na porta definida

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer server.Close() // Garantindo que a conexão vai ser encerrada

	fmt.Println("Esperando conexão em " + CONN_HOST + ":" + CONN_PORT) //Indicativo de inicizalização do server

	for {
		conn, err := server.Accept() // Em loop esperando alguma conexão a ser aceita
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn) // Tratando cada conexão em thread
	}

}

func handleRequest(connBrowser net.Conn) {

	fmt.Println("==============================NAVEGADOR CONECTADO==============================")

	defer connBrowser.Close()           // Garantindo que a conexão será finalizada
	var resultCliente [10000]byte       // Array para receber a requisição
	connBrowser.Read(resultCliente[0:]) // Lendo a requisição do navegador

	url := getURLNavegador(string(resultCliente[0:])) //==========================================
	nome, diretorio := separaURL(url)                 //Extraindo informações vindas da requisição
	tempo := getTime()                                //==========================================

	noBuscado := procuraCache(nome, diretorio) //Verificando se existe registro igual na cache

	if noBuscado != nil { //ENCONTRAMOS A MESMA REQUISIÇÃO NA MEMÓRIA CACHE

		fmt.Println("==============================RECUPERANDO DA CACHE==============================")
		fmt.Println("URL REQUISITADA: " + noBuscado.url)
		fmt.Println("HOST: " + noBuscado.nome)
		fmt.Println("CAMINHO: " + noBuscado.diretorio)
		fmt.Println("ENDERECO IP: " + noBuscado.ip)
		fmt.Print("ACESSADO EM: ")
		fmt.Println(noBuscado.tempo)

		param, _ := strconv.Atoi(os.Args[1])

		if (noBuscado.tempo + param) > getTime() { //Se está dentro do limite de tempo retorna o mesmo conteudo

			fmt.Println("==============================AINDA ESTÁ VÁLIDO==============================")

			connBrowser.Write([]byte(noBuscado.content))

		} else { //Se expirou, usa a conexão externa no servidor que contém a resposta

			fmt.Println("==============================NÃO ESTÁ VÁLIDO==============================")

			conexaoExterna(noBuscado, connBrowser, 0)

		}
	} else {

		novo := no{ //===============================================================
			url:       url,       //===============================================================
			nome:      nome,      // Criando um novo elemento para a cache e definindo seus valores
			diretorio: diretorio, //===============================================================
			tempo:     tempo,     //===============================================================
		}

		novo.ip = urlParaIp(novo.nome) // Convertendo o endereço URL em ip

		fmt.Println("==============================CRIANDO NOVO REGISTRO==============================")
		fmt.Println("URL REQUISITADA: " + novo.url)
		fmt.Println("HOST: " + novo.nome)
		fmt.Println("CAMINHO: " + novo.diretorio)
		fmt.Println("ENDERECO IP: " + novo.ip)
		fmt.Print("ACESSADO EM: ")
		fmt.Println(novo.tempo)

		conexaoExterna(&novo, connBrowser, 1)
	}
	gravarCache() //grava tudo que está no vetor CACHE no arquivo "cache.txt"
}

func conexaoExterna(node *no, connBrowser net.Conn, key int) {
	if node.ip != "" {

		fmt.Println("==============================CONEXÃO EXTERNA==============================")

		serviceRequisitado := node.ip + ":80"                        //novoNO.ip + ":80"
		tcpAddr, _ := net.ResolveTCPAddr("tcp4", serviceRequisitado) //Conectando ao server externo
		connServer, _ := net.DialTCP("tcp", nil, tcpAddr)            //Conectando ao server externo

		requisicao := "GET " + node.diretorio + " HTTP/1.0\r\n\r\n" //Montando a requisição

		_, _ = connServer.Write([]byte(requisicao)) //Requisitando ao server externo

		resultServer, _ := ioutil.ReadAll(connServer) //Lendo tudo que vem do server externo

		connBrowser.Write([]byte(resultServer[0:])) //Enviando a resposta ao navegador
		node.content = resultServer                 // Salvando no elemento da cache o conteudo retornado

		if key == 1 {
			cache = append(cache, *node) //Inserindo na cache
		} else {
			node.tempo = getTime() //Atuzalizando tempo de um registro que já estava na cache
		}
	} else {
		fmt.Println("==============================NAO GERA IP VÁLIDO==============================")
	}

	connBrowser.Close() //Encerrando a conexão com o navegador

	fmt.Println("==============================ESTADO ATUAL DA CACHE==============================")
	for i := 0; i < len(cache); i++ {
		fmt.Println("URL REQUISITADA: " + cache[i].url)
		fmt.Println("HOST: " + cache[i].nome)
		fmt.Println("CAMINHO: " + cache[i].diretorio)
		fmt.Println("ENDERECO IP: " + cache[i].ip)
		fmt.Print("ACESSADO EM: ")
		fmt.Println(cache[i].tempo)
		fmt.Println("==========================================================================================")
	}
	fmt.Println("==============================ENCERRANDO A CONEXAO==============================\n\n\n\n\n ")

	gravarCache() //grava tudo que está no vetor CACHE no arquivo "cache.txt"
}

func procuraCache(nome string, diretorio string) *no {
	for i := 0; i < len(cache); i++ {
		if nome == cache[i].nome && diretorio == cache[i].diretorio {
			return &cache[i]
		}
	}
	return nil
}

func getURLNavegador(str string) string {
	end := ""
	i := 5
	for i < len(str)+1 {
		if string(str[i]) == " " || i == (len(str)-1) {
			break
		}
		end = end + string(str[i])
		i++
	}

	return end
}

func separaURL(str string) (string, string) {
	end := ""
	dir := "/"

	i := 0
	for i < len(str) {
		if string(str[i]) == "/" {
			i++
			break
		}
		end = end + string(str[i])
		i++
	}

	for i < len(str) {
		dir = dir + string(str[i])
		i++
	}
	//fmt.Println(end)
	//fmt.Println(dir)
	return end, dir
}

func urlParaIp(url string) string {
	addr, err := net.ResolveIPAddr("ip", url)
	if err != nil {
		return ""
	}
	return addr.String()
}

func getTime() int {
	// get the location
	location, _ := time.LoadLocation("Europe/Rome")

	// this should give you time in location
	t := time.Now().In(location)

	tempo := t.String()

	convert := ""

	for i := 11; i < 19; i++ {
		convert = convert + string(tempo[i])
	}

	segundos, _ := strconv.Atoi(convert[6:])
	minutos, _ := strconv.Atoi(convert[3:5])
	horas, _ := strconv.Atoi(convert[0:2])

	totalEmSegundos := segundos + (60 * minutos) + (3600 * horas)
	return totalEmSegundos
}

func carregarCache() {
	leitura, err := lerArquivo("cache.txt") //Lê o que tem no arquivo
	if err != nil {
		log.Fatalf("Erro:", err)
	}

	i := 0
	for {
		if len(leitura) == 0 {
			break
		}
		novo := no{
			url:       leitura[i],
			nome:      leitura[i+1],
			diretorio: leitura[i+2],
			ip:        leitura[i+3],
		}
		t, _ := strconv.Atoi(leitura[i+4])
		novo.tempo = t

		i += 1

		//obtendo o conteúdo
		var conteudo []byte
		for {
			if leitura[i] == "FIM_REGISTRO" {
				i += 1
				break
			}
			linhaConvertida := []byte(leitura[i])
			for j := 0; j < len(linhaConvertida); j++ {
				conteudo = append(conteudo, linhaConvertida[j])
			}
			i += 1
		}

		novo.content = conteudo

		cache = append(cache, novo)

		//i += 5

		if i >= len(leitura) {
			break
		}
	}
}

func gravarCache() {
	var escrita []string
	if len(cache) != 0 {
		for i := 0; i < len(cache); i++ {
			escrita = append(escrita, cache[i].url)
			escrita = append(escrita, cache[i].nome)
			escrita = append(escrita, cache[i].diretorio)
			escrita = append(escrita, cache[i].ip)
			escrita = append(escrita, strconv.Itoa(cache[i].tempo))
			escrita = append(escrita, string(cache[i].content))
			escrita = append(escrita, "FIM_REGISTRO")
		}
		escreverArquivo(escrita, "cache.txt")
	}
}

func lerArquivo(caminhoDoArquivo string) ([]string, error) { // OK!

	arquivo, err := os.Open(caminhoDoArquivo) // Abre o arquivo
	if err != nil {                           // Caso tenha encontrado algum erro ao tentar abrir o arquivo retorne o erro encontrado
		return nil, err
	}
	defer arquivo.Close() // Garante que o arquivo sera fechado apos o uso

	var linhas []string
	scanner := bufio.NewScanner(arquivo) // Cria um scanner que lê cada linha do arquivo
	for scanner.Scan() {
		linhas = append(linhas, scanner.Text())
	}

	return linhas, scanner.Err() // Retorna as linhas lidas e um erro se ocorrer algum erro no scanner
}

func escreverArquivo(linhas []string, caminhoDoArquivo string) error {

	arquivo, err := os.Create(caminhoDoArquivo) // Cria o arquivo de texto. se já existir, apenas abre
	if err != nil {                             // Caso tenha encontrado algum erro retornar ele
		return err
	}
	defer arquivo.Close() // Garante que o arquivo sera fechado apos o uso

	escritor := bufio.NewWriter(arquivo) // Cria um escritor responsavel por escrever cada linha do slice no arquivo de texto
	for _, linha := range linhas {
		fmt.Fprintln(escritor, linha)
	}

	return escritor.Flush() // Caso a funcao flush retorne um erro ele sera retornado aqui tambem
}

func printArquivo() {
	fmt.Println("<<<<<< CONTEÚDO DO ARQUIVO CACHE.TXT >>>>>>")
	conteudo, err := lerArquivo("cache.txt") //leitura do arquivo
	if err != nil {                          //verifica algum erro
		log.Fatalf("Erro:", err)
	}

	for indice, linha := range conteudo { //printa índice da linha e conteúdo
		fmt.Println(indice, linha)
	}
}

//localhost:8888/www-net.cs.umass.edu/personnel/kurose.html
//localhost:8888/www.example.org
//localhost:8888/www.google.com
//localhost:8888/si3.ufc.br/sigaa/verTelaLogin.do
