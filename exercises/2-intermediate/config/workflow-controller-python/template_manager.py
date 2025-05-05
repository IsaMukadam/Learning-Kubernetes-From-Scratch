import yaml
import os

class TemplateManager:
    def __init__(self):
        self.templates = {}

    def load_templates(self, template_dir):
        """Load task templates from a directory"""
        for filename in os.listdir(template_dir):
            if filename.endswith('.yaml'):
                with open(os.path.join(template_dir, filename)) as f:
                    template = yaml.safe_load(f)
                    if template and 'name' in template:
                        self.templates[template['name']] = template

    def get_template(self, template_name):
        """Get a template by name"""
        return self.templates.get(template_name)

    def render_template(self, template_name, variables):
        """Render a template with variables"""
        template = self.get_template(template_name)
        if not template:
            raise ValueError(f"Template not found: {template_name}")

        # Deep copy the template to avoid modifying the original
        import copy
        rendered = copy.deepcopy(template)

        # Replace variables in the template
        self._replace_variables(rendered, variables)
        return rendered

    def _replace_variables(self, obj, variables):
        """Recursively replace variables in an object"""
        if isinstance(obj, dict):
            for key, value in obj.items():
                if isinstance(value, (dict, list)):
                    self._replace_variables(value, variables)
                elif isinstance(value, str):
                    for var_name, var_value in variables.items():
                        placeholder = f"${{{var_name}}}"
                        if placeholder in value:
                            obj[key] = value.replace(placeholder, str(var_value))
        elif isinstance(obj, list):
            for i, item in enumerate(obj):
                if isinstance(item, (dict, list)):
                    self._replace_variables(item, variables)
                elif isinstance(item, str):
                    for var_name, var_value in variables.items():
                        placeholder = f"${{{var_name}}}"
                        if placeholder in item:
                            obj[i] = item.replace(placeholder, str(var_value))