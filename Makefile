GOCMD := go

FRONTEND_DIR := ./front-end
INDEXJS := $(FRONTEND_DIR)/dist/index.js
LEMIN := ./lem-in
VISUALISER := ./visualiser

.PHONY: all
all: $(LEMIN) $(VISUALISER)

$(LEMIN):
	$(GOCMD) build -o $@ ./cmd

$(VISUALISER): $(INDEXJS)
	$(GOCMD) build -o visualiser $(FRONTEND_DIR)/src

$(INDEXJS):
	cd $(FRONTEND_DIR) && npm install && npm run build

.PHONY: clean
clean:
	$(RM) $(LEMIN) $(VISUALISER)
	$(RM) -r $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules
