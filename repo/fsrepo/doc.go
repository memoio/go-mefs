// package fsrepo
//
// TODO explain the package roadmap...
//
//   .mefs/
//   ├── client/
//   |   ├── client.lock          <------ protects client/ + signals its own pid
//   │   ├── mefs-client.cpuprof
//   │   └── mefs-client.memprof
//   ├── config
//   ├── daemon/
//   │   ├── daemon.lock          <------ protects daemon/ + signals its own address
//   │   ├── mefs-daemon.cpuprof
//   │   └── mefs-daemon.memprof
//   ├── datastore/
//   ├── repo.lock                <------ protects datastore/ and config
//   └── version
package fsrepo

// TODO prevent multiple daemons from running
