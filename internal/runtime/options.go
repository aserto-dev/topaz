package runtime

import (
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage"
)

type Option func(*Runtime)

func WithPlugin(name string, factory plugins.Factory) Option {
	return func(r *Runtime) {
		r.plugins[name] = factory
	}
}

func WithBuiltin1(decl *rego.Function, impl rego.Builtin1) Option {
	return func(r *Runtime) {
		r.builtins1[decl] = impl
		r.builtins = append(r.builtins, rego.Function1(decl, impl))
		r.compilerBuiltins[decl.Name] = &ast.Builtin{
			Name: decl.Name,
			Decl: decl.Decl,
		}
	}
}

func WithBuiltin2(decl *rego.Function, impl rego.Builtin2) Option {
	return func(r *Runtime) {
		r.builtins2[decl] = impl
		r.builtins = append(r.builtins, rego.Function2(decl, impl))
		r.compilerBuiltins[decl.Name] = &ast.Builtin{
			Name: decl.Name,
			Decl: decl.Decl,
		}
	}
}

func WithBuiltin3(decl *rego.Function, impl rego.Builtin3) Option {
	return func(r *Runtime) {
		r.builtins3[decl] = impl
		r.builtins = append(r.builtins, rego.Function3(decl, impl))
		r.compilerBuiltins[decl.Name] = &ast.Builtin{
			Name: decl.Name,
			Decl: decl.Decl,
		}
	}
}

func WithBuiltin4(decl *rego.Function, impl rego.Builtin4) Option {
	return func(r *Runtime) {
		r.builtins4[decl] = impl
		r.builtins = append(r.builtins, rego.Function4(decl, impl))
		r.compilerBuiltins[decl.Name] = &ast.Builtin{
			Name: decl.Name,
			Decl: decl.Decl,
		}
	}
}

func WithBuiltinDyn(decl *rego.Function, impl rego.BuiltinDyn) Option {
	return func(r *Runtime) {
		r.builtinsDyn[decl] = impl
		r.builtins = append(r.builtins, rego.FunctionDyn(decl, impl))
		r.compilerBuiltins[decl.Name] = &ast.Builtin{
			Name: decl.Name,
			Decl: decl.Decl,
		}
	}
}

func WithStorage(storageInterface storage.Store) Option {
	return func(r *Runtime) {
		r.storage = storageInterface
	}
}

func WithImport(imp string) Option {
	return func(r *Runtime) {
		r.imports = append(r.imports, imp)
	}
}

func WithImports(imp []string) Option {
	return func(r *Runtime) {
		r.imports = append(r.imports, imp...)
	}
}

func WithRegoVersion(v ast.RegoVersion) Option {
	return func(r *Runtime) {
		r.regoVersion = v
	}
}
