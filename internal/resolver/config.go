package resolver

import (
	"context"
	"fmt"

	cloudproviderv1 "git.cells.es/mis/visa/visa-k8s-controller/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResolvedConfig struct {
	Image          *cloudproviderv1.VISACloudImage
	Flavour        *cloudproviderv1.VISACloudFlavour
	Sidecars       []*cloudproviderv1.VISASidecarConfig
	SecurityGroups []*cloudproviderv1.VISACloudSecurityGroup
}

type ConfigResolver struct {
	client client.Client
}

func (c *ConfigResolver) get(
	ctx context.Context,
	namespace string, name string,
	obj client.Object,
) error {
	return c.client.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, obj)
}

func (c *ConfigResolver) getImage(ctx context.Context, namespace string, name string) (*cloudproviderv1.VISACloudImage, error) {
	image := &cloudproviderv1.VISACloudImage{}
	if err := c.get(ctx, namespace, name, image); err != nil {
		return nil, fmt.Errorf("getting cloud image %q: %w", name, err)
	}

	return image, nil
}

func (c *ConfigResolver) getFlavour(ctx context.Context, namespace string, name string) (*cloudproviderv1.VISACloudFlavour, error) {
	flavour := &cloudproviderv1.VISACloudFlavour{}
	if err := c.get(ctx, namespace, name, flavour); err != nil {
		return nil, fmt.Errorf("getting cloud flavour %q: %w", name, err)
	}

	return flavour, nil
}

func (c *ConfigResolver) getSidecarConfigs(ctx context.Context, namespace string, sidecars []cloudproviderv1.SidecarConfigRef) ([]*cloudproviderv1.VISASidecarConfig, error) {

	sidecarConfigs := make(
		[]*cloudproviderv1.VISASidecarConfig,
		0,
		len(sidecars))

	for _, sidecarRef := range sidecars {
		name := sidecarRef.Name
		sidecar := &cloudproviderv1.VISASidecarConfig{}
		if err := c.get(ctx, namespace, name, sidecar); err != nil {
			return nil, fmt.Errorf("getting cloud sidecar %q: %w", name, err)
		}
		sidecarConfigs = append(sidecarConfigs, sidecar)
	}

	return sidecarConfigs, nil
}

func (c *ConfigResolver) getSecurityGroups(ctx context.Context, namespace string, groups []cloudproviderv1.LocalObjectReference) ([]*cloudproviderv1.VISACloudSecurityGroup, error) {
	securityGroups := make(
		[]*cloudproviderv1.VISACloudSecurityGroup,
		0,
		len(groups))

	for _, secGroupRef := range groups {
		name := secGroupRef.Name
		group := &cloudproviderv1.VISACloudSecurityGroup{}
		if err := c.get(ctx, namespace, name, group); err != nil {
			return nil, fmt.Errorf("getting cloud security group %q: %w", name, err)
		}
		securityGroups = append(securityGroups, group)
	}

	return securityGroups, nil
}

func New(client client.Client) *ConfigResolver {
	return &ConfigResolver{
		client: client,
	}
}

func (c *ConfigResolver) ResolveConfiguration(ctx context.Context, visaInstance *cloudproviderv1.VISAInstance) (*ResolvedConfig, error) {
	image, err := c.getImage(ctx, visaInstance.Namespace, visaInstance.Spec.Image.Name)
	if err != nil {
		return nil, err
	}

	flavour, err := c.getFlavour(ctx, visaInstance.Namespace, visaInstance.Spec.Flavour.Name)
	if err != nil {
		return nil, err
	}

	sidecars, err := c.getSidecarConfigs(ctx, visaInstance.Namespace, flavour.Spec.Sidecars)
	if err != nil {
		return nil, err
	}

	securityGroups, err := c.getSecurityGroups(ctx, visaInstance.Namespace, visaInstance.Spec.SecurityGroups)
	if err != nil {
		return nil, err
	}

	return &ResolvedConfig{
		Image:          image,
		Flavour:        flavour,
		Sidecars:       sidecars,
		SecurityGroups: securityGroups,
	}, nil
}
