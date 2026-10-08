# ADR 001: frontend sem framework no protótipo

- **Status:** aceito
- **Data:** 2026-10-08

## Contexto

O GoPing precisa de uma interface de portfólio responsiva, indexável e pronta para evoluir junto de um backend escrito em Go. O projeto ainda não possui API nem requisitos que exijam gerenciamento de estado complexo.

## Decisão

Usar HTML semântico, CSS e JavaScript do navegador no primeiro protótipo. Dados demonstrativos são persistidos em `localStorage`. Ícones são carregados pela biblioteca Lucide via CDN.

## Consequências

### Positivas

- Nenhuma etapa de build ou dependência de runtime local.
- Conteúdo principal presente no HTML para rastreamento e acessibilidade.
- Migração direta para templates Go no MVP.
- Menor superfície de manutenção durante o aprendizado do backend.

### Limitações

- O protótipo não executa verificações reais.
- O CDN de ícones é uma dependência externa e deve ser empacotado antes de um deploy com política CSP restrita.
- Componentes e estado exigirão nova avaliação se a interface crescer significativamente.

## Gatilhos para revisão

Reavaliar quando houver rotas independentes, estado compartilhado complexo, equipe frontend separada ou necessidade comprovada de renderização híbrida. A adoção de um framework não deve ocorrer apenas por preferência tecnológica.
