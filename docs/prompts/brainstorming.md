# Brainstorming with an AI Agent Before Generating Step-by-Step Guides

When working with an AI agent (such as GitHub Copilot) to generate a step-by-step guide from a phased development plan, it's valuable to have a brainstorming session first. This helps ensure the agent understands your intentions, clarifies any ambiguities, and aligns the output with your needs.

---

## Why Brainstorm First?

- **Clarify intent:** The AI can ask questions or point out potential ambiguities.
- **Surface assumptions:** You and the agent can discuss any implicit expectations (e.g., single vs. multiple instances, local vs. cloud).
- **Align on scope:** Ensure the guide matches the incremental approach of your plan.
- **Avoid mistakes:** Prevent the AI from defaulting to general best practices when your plan requires something different.

---

## Suggested Brainstorming Workflow

### 1. Start with an Open-Ended Prompt

Invite the agent to analyze your development plan and ask questions before writing the guide.

> I want to generate a step-by-step guide for phase 5 of my development plan. Before doing so, let’s brainstorm what the intent and requirements for phase 5 are. Please analyze the plan, identify any ambiguities or decisions to be made (such as number of instances, local vs. cloud deployment), and ask me clarifying questions before proceeding.

---

### 2. Encourage Critical Thinking

Ask the agent to compare general best practices with your plan, and highlight any differences.

> Please also let me know if there are best practices that might differ from the plan so we can discuss whether to follow the plan strictly or adapt it.

---

### 3. Allow Back-and-Forth Clarification

Be prepared to answer the agent’s questions about your goals and the specifics of your plan. Clarify points like:
- Should this phase be local only or also cloud-ready?
- Should we use just one replica per service, or more?
- Are high availability and scaling needed now, or in a later phase?

---

### 4. Summarize the Agreed Approach

Once you’ve resolved all ambiguities, ask the agent to summarize what’s been agreed on.

> Based on our discussion, please summarize your understanding of the requirements for phase 5. Once I confirm, generate the step-by-step guide.

---

### 5. Proceed to Guide Generation

With alignment achieved, instruct the agent to generate the step-by-step guide according to the clarified requirements.

---

## Sample Brainstorming Prompt

> Let’s brainstorm before generating a deployment guide for phase 5 of my development plan.
> Please:
> - Analyze the plan and outline any key decisions or ambiguities.
> - List questions or clarifications you need from me (e.g., number of instances, local vs. cloud).
> - Point out where general Kubernetes best practices might differ from the plan.
    > Once we’ve clarified everything, summarize your understanding and wait for my confirmation before generating the guide.

---

**Tip:** This approach is especially useful if you’re not deeply familiar with the deployment technology (like Kubernetes) and want to make sure the AI doesn’t make incorrect assumptions.
