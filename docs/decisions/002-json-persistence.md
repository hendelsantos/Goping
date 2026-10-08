# ADR 002: persistência JSON antes do SQLite

- **Status:** aceito temporariamente
- **Data:** 2026-10-08

## Contexto

O primeiro backend precisa demonstrar API, validação, concorrência e persistência sem antecipar migrations e SQL antes do checker HTTP.

## Decisão

Persistir monitores e atividades em um único arquivo JSON. O armazenamento usa `sync.RWMutex`, grava em arquivo temporário e publica com renomeação atômica.

## Consequências

- O projeto continua sem dependências externas e simples de executar.
- Reinicializações não apagam cadastros.
- Apenas uma instância do processo pode escrever com segurança no arquivo.
- Consultas, histórico volumoso e transações complexas não são adequados a esse formato.

## Gatilho de substituição

Migrar para SQLite antes de implementar histórico de verificações e scheduler persistente. A interface do armazenamento e os contratos HTTP devem preservar o frontend durante a troca.
