<template>
    <nav aria-label="Breadcrumb">
        <div class="ui breadcrumb">
            <template v-for="(item, index) in items" :key="index">
                <div v-if="index !== 0" class="divider"> / </div>

                <a
                    v-if="!item.active"
                    class="section"
                    :href="item.href || '#'"
                    @click.prevent="onClick(item)"
                >
                    {{ item.text }}
                </a>

                <div v-else class="active section">{{ item.text }}</div>
            </template>
        </div>
    </nav>
</template>

<script setup lang="ts">
interface BreadcrumbItem {
    text: string
    href?: string
    active?: boolean
}

defineProps<{
    items?: BreadcrumbItem[]
}>()

const emit = defineEmits<{
    navigate: [item: BreadcrumbItem]
}>()

function onClick(item: BreadcrumbItem) {
    emit('navigate', item)
}
</script>

<style scoped>
.ui.breadcrumb {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    font-size: 14px;
    padding: 8px;
}
.ui.breadcrumb .divider {
    margin: 0 0.5em;
    color: rgba(0, 0, 0, 0.6);
}
.ui.breadcrumb .section {
    color: #4183c4;
    text-decoration: none;
    cursor: pointer;
}
.ui.breadcrumb .active.section {
    color: rgba(0, 0, 0, 0.85);
    cursor: default;
    font-weight: 600;
}
</style>