library(dplyr)
library(ggplot2)
library(tidyr)
library(stringr)

go_timings <- read.csv("timings_go.csv")
python_timings <- read.csv("timings_python.csv")

go_timings$Language = "Go"
python_timings$Language = "Python"
timings <- rbind(go_timings, python_timings) 
# convert data to long format
timings <- timings %>% 
  pivot_longer(
    c("Attach", "Single.Step", "Breakpoint", "Registers", "Memory", "Memory.Search", "Total"),
    names_to = "Functionality",
    values_to = "Time",
  ) %>% 
  mutate(Language = as.factor(Language)) %>%
  as.data.frame()

timings$Functionality <- sapply(timings$Functionality, function(s) {
  str_replace(s, "\\.", " ")
})
names(timings$Functionality) = NULL
timings <- timings %>% mutate(Functionality = as.factor(Functionality))

median_timings <- timings %>% 
  group_by(Functionality, Language) %>% 
  summarise(Median.Time = median(Time)) %>% 
  as.data.frame()
write.csv(median_timings, "median_timings.csv", row.names = FALSE)

plot <- timings %>%
  ggplot() +
  aes(x = Functionality, y = Time, color = Language) +
  geom_boxplot() +
  xlab("Functionality") +
  ylab("Time (ms)")
ggsave("analysis.png", plot, width=7, height=4)

Functionality = levels(timings$Functionality)

columns = c("Functionality", "Difference estimate", "95-percent confidence interval",  "p-value")
stat_results = data.frame(matrix(nrow = 0, ncol = length(columns)))

for (f in Functionality) {
  filtered_times <- timings %>% filter(Functionality == f)
  row <- tryCatch({
    r <- wilcox.test(Time ~ Language, data = filtered_times, conf.int = T)
    confInt <- paste(sprintf("%0.5f", r$conf.int[1]), "to",  sprintf("%0.5f", r$conf.int[2]))
    c(f, sprintf("%0.5f", r$estimate), confInt, formatC(r$p.value, format = "e", digits = 2))
  }, error = function(e) {
    c(f, NA, NA, NA)
  })
  stat_results <- rbind(stat_results, row)
}

colnames(stat_results) = columns
write.csv(stat_results, "stat_analysis.csv", row.names = FALSE)
