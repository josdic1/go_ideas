async function loadIdeas() {
  const res = await fetch("/ideas");
  const ideas = await res.json();

  const list = document.getElementById("ideas");
  list.innerHTML = "";

  ideas.forEach((i) => {
    const li = document.createElement("li");
    li.textContent = i.idea;
    list.appendChild(li);
  });
}

async function createIdea() {
  const input = document.getElementById("ideaInput");
  const idea = input.value;

  await fetch("/ideas", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ idea }),
  });

  input.value = "";

  loadIdeas();
}

document.getElementById("addBtn").onclick = createIdea;

loadIdeas();
