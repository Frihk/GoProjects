FRONTEND_DIR := ./front-end
GOCMD := go

.PHONY: all
all: lem-in visualiser

.PHONY: lem-in
lem-in:
	$(GOCMD) build -o lem-in ./cmd

.PHONY: frontend
frontend:
	cd $(FRONTEND_DIR) && npm install && npm run build

.PHONY: visualiser
visualiser: frontend
	$(GOCMD) build -o visualiser $(FRONTEND_DIR)/src

.PHONY: clean
clean:
	rm -f lem-in visualiser
	rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules
