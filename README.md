# Betstake Webscrap

Projeto em Go para realizar raspagem web (web scraping) e automação de dados de partidas de futebol (com foco em E-Soccer tipo GT Leagues) em plataformas de apostas, utilizando a biblioteca [go-rod](https://github.com/go-rod/rod).

## 🚀 Funcionalidades Principais (O que o projeto faz)

1. **Inicialização e Bypass de Modais (`browser.LoadPageFlow`)**
   - Inicia uma instância do navegador Chromium (modo Headless ou Visível).
   - Navega até a página de eventos esportivos ao vivo.
   - Aceita e fecha automaticamente avisos de maioridade (+18) e banners de consentimento de Cookies.

2. **Autenticação Automática (`auth.LoginFlow`)** *(Opcional)*
   - Simula a interação com o painel de login na plataforma.
   - Pede os dados de acesso e lida com verificações de duas etapas (2FA) quando exigido.

3. **Busca Dinâmica de Ligas (`crawler.FindLeague` e `ExpandLeague`)**
   - Realiza scroll automático na página (forçando o *Lazy Load* de elementos visuais) até localizar a **"GT Leagues"**.
   - Faz busca dentro da árvore DOM principal e também dentro de `iframes`.
   - Expande os componentes colapsados dessa liga específica.

4. **Raspagem de Partidas (`crawler.GetMatches`)**
   - Extrai dos jogos listados: Nome dos Times, Placar e o Tempo da Partida.
   - Evita dados duplicados analisando strings das chaves das equipes.

5. **Exame Profundo de Mercados de Handicap**
   - Clica especificamente no título da partida (nomes das equipes) para carregar todos os mercados detalhados daquele jogo.
   - Realiza *scroll* automático pelo painel interno de detalhes buscando o mercado alvo.
   - Assim que acha o painel de **"1st Half Asian Handicap"** (ou variações em pt-br), ele abre o mercado e mapeia todas as Linhas (`Line`) e Cotações (`Odds`) em tempo real.

## 📋 Informações que o projeto solicita / requer

Para operar corretamente o robô precisa que você forneça/valide:

- **Credenciais de Acesso (Login):** Caso o login esteja ativado (`main.go`), o programa pedirá no seu Terminal o preenchimento manual do seu **Usuário** e **Senha**. Se atrelado à conta, pedirá também sua **Chave de 2FA/PIN**.
- **Sessão Local Persistente:** O projeto precisa do diretório (`./user-data`) para armazenar os *cookies e o seu cache de navegação*. Isso evita que o robô tenha que ser logado a todo instante.
- **Configurações de Alvos (Hardcoded):** Atualmente, o bot busca especificamente por:
  - Textos que indicam a liga-alvo *(ex. "GT Leagues")*
  - O mercado de *(ex. "1st Half Asian Handicap")*
  - Caso os botões e os identificadores HTML da plataforma de aposta mudem, os seletores nos arquivos `.go` devem ser atualizados.

## 💻 Como Executar na Máquina Local

Instale as dependências e rode o script:

```bash
go mod tidy
go run main.go
```
