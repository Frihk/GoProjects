GOCMD := go

FRONTEND_DIR := ./front-end
INDEXJS := $(FRONTEND_DIR)/dist/index.js
LEMIN := ./lem-in
VISUALISER := ./visualiser

.PHONY: all
all: $(LEMIN) $(VISUALISER)

.PHONY: $(LEMIN)
$(LEMIN):
	$(GOCMD) build -o $@ ./cmd

.PHONY: $(VISUALISER)
$(VISUALISER): $(INDEXJS)
	$(GOCMD) build -o visualiser $(FRONTEND_DIR)/src

.PHONY: $(INDEXJS)
$(INDEXJS):
	cd $(FRONTEND_DIR) && npm install && npm run build

.PHONY: clean
clean:
	$(RM) $(LEMIN) $(VISUALISER)
	$(RM) -r $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules
