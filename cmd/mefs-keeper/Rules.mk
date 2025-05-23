include mk/header.mk
MEFS_BIN_$(d) := $(call go-curr-pkg-tgt)

TGT_BIN += $(MEFS_BIN_$(d))
TEST_GO_BUILD += $(d)-try-build
CLEAN += $(MEFS_BIN_$(d))

PATH := $(realpath $(d)):$(PATH)

# disabled for now
# depend on *.pb.go files in the repo as Order Only (as they shouldn't be rebuilt if exist)
# DPES_OO_$(d) := diagnostics/pb/diagnostics.pb.go exchange/bitswap/message/pb/message.pb.go
# DEPS_OO_$(d) += merkledag/pb/merkledag.pb.go namesys/pb/namesys.pb.go
# DEPS_OO_$(d) += pin/internal/pb/header.pb.go unixfs/pb/unixfs.pb.go

$(d)_flags =-ldflags="-X "github.com/memoio/go-mefs".CurrentCommit=$(git-hash)"

$(d)-try-build $(MEFS_BIN_$(d)): GOFLAGS += $(cmd/mefs_flags)

# uses second expansion to collect all $(DEPS_GO)
$(MEFS_BIN_$(d)): $(d) $$(DEPS_GO) ALWAYS #| $(DEPS_OO_$(d))
	$(go-build)

TRY_BUILD_$(d)=$(addprefix $(d)-try-build-,$(SUPPORTED_PLATFORMS))
$(d)-try-build: $(TRY_BUILD_$(d))
.PHONY: $(d)-try-build

$(TRY_BUILD_$(d)): PLATFORM = $(subst -, ,$(patsubst $<-try-build-%,%,$@))
$(TRY_BUILD_$(d)): GOOS = $(word 1,$(PLATFORM))
$(TRY_BUILD_$(d)): GOARCH = $(word 2,$(PLATFORM))
$(TRY_BUILD_$(d)): $(d) $$(DEPS_GO) ALWAYS
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(go-try-build)
.PHONY: $(TRY_BUILD_$(d))

$(d)-install: GOFLAGS += $(cmd/mefs-keeper_flags)
$(d)-install: $(d) $$(DEPS_GO) ALWAYS 
	$(GOCC) install $(go-flags-with-tags) ./cmd/mefs-keeper
.PHONY: $(d)-install

COVER_BIN_$(d) := $(d)/mefs-test-cover
CLEAN += $(COVER_BIN_$(d))

$(COVER_BIN_$(d)): GOTAGS += testrunmain
$(COVER_BIN_$(d)): $(d) $$(DEPS_GO) ALWAYS
	$(eval TMP_PKGS := $(shell $(GOCC) list -f '{{range .Deps}}{{.}} {{end}}' $(go-flags-with-tags) ./cmd/mefs-keeper | sed 's/ /\n/g' | grep mefs/go-mefs) $(call go-pkg-name,$<))
	$(eval TMP_LIST := $(call join-with,$(comma),$(TMP_PKGS)))
	@echo $(GOCC) test $@ -c -covermode atomic -coverpkg ... $(go-flags-with-tags) ./$(@D) # for info
	@$(GOCC) test -o $@ -c -covermode atomic -coverpkg $(TMP_LIST) $(go-flags-with-tags) ./$(@D) 2>&1 | (grep -v 'warning: no packages being tested' || true)

include mk/footer.mk
