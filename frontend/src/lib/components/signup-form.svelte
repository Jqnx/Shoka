<script lang="ts">
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { superForm, type Infer, type SuperValidated } from 'sveltekit-superforms';
	import { signupSchema, type SignupSchema } from '$lib/schemas/signup';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import Label from './ui/label/label.svelte';

	let { data }: { data: { form: SuperValidated<Infer<SignupSchema>> } } = $props();

	// svelte-ignore state_referenced_locally
	const form = superForm(data.form, {
		validators: zod4Client(signupSchema)
	});

	const { form: formData, enhance } = form;
</script>

<Card.Root class="mx-auto w-full max-w-sm">
	<Card.Header>
		<Card.Title>Create an account</Card.Title>
		<Card.Description>Enter your information below to create your account</Card.Description>
	</Card.Header>
	<Card.Content>
		<form method="POST" use:enhance>
			<Form.Field {form} name="username">
				<Form.Control>
					{#snippet children({ props })}
						<Form.Label>Username</Form.Label>
						<Input {...props} bind:value={$formData.username} />
					{/snippet}
				</Form.Control>
				<Form.FieldErrors />
			</Form.Field>
			<Form.Field {form} name="email">
				<Form.Control>
					{#snippet children({ props })}
						<Form.Label>Email</Form.Label>
						<Input {...props} bind:value={$formData.email} />
					{/snippet}
				</Form.Control>
				<Form.FieldErrors />
			</Form.Field>
			<Form.Field {form} name="password">
				<Form.Control>
					{#snippet children({ props })}
						<Form.Label>Password</Form.Label>
						<Input {...props} bind:value={$formData.password} type="password" />
					{/snippet}
				</Form.Control>
				<Form.Description>Must be at least 8 characters long.</Form.Description>
				<Form.FieldErrors />
			</Form.Field>
			<Form.Field {form} name="confirmPassword">
				<Form.Control>
					{#snippet children({ props })}
						<Form.Label>Confirm Password</Form.Label>
						<Input {...props} bind:value={$formData.confirmPassword} type="password" />
					{/snippet}
				</Form.Control>
				<Form.FieldErrors />
			</Form.Field>
			<Form.Button class="w-full">Sign Up</Form.Button>
			<Label
				class="justify-self-center pt-2 text-center text-sm leading-normal font-normal text-muted-foreground"
			>
				Already have an account? <a href={resolve('/login')} class="underline">Sign in</a>
			</Label>
		</form>
	</Card.Content>
</Card.Root>
