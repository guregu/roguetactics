def domains;

new File("strings.csv").splitEachLine(",") {fields ->
  people.add(
    if (!domains[fields[0]]) {
    	domains[fields[0]] = 1;
    } else {
    	domains[fields[0]] = domains[fields[0]]++;
    }
  )
}
